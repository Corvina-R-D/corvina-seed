package seed

import (
	"context"
	"corvina/corvina-seed/src/seed/api"
	"corvina/corvina-seed/src/seed/dto"
	"corvina/corvina-seed/src/utils"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

func Execute(ctx context.Context, input *dto.ExecuteInDTO) error {

	totalStart := time.Now()

	organization, err := api.GetOrganizationMine(ctx)
	if err != nil {
		return err
	}

	log.Info().Interface("organization", organization).Msg("Organization retrieved")

	// rootOrg è sempre la org dell'admin (es. "exor"): ha il realm Keycloak
	// configurato ed è l'unica che accetta la creazione di utenti con email @realm.
	// Le sub-org create dal seed tool non hanno hostname e rifiutano la creazione
	// utenti (errore 15004). Passiamo rootOrg a createUsers esplicitamente.
	rootOrg := organization

	// Se è stato fornito --seed-org-name, creiamo una sub-org con quel nome
	// come "contenitore" di tutto ciò che segue (albero, device e sub-org).
	// L'org radice dell'admin viene usata solo come parent di questa nuova org.
	// Al termine stampiamo il resource_id così può essere riutilizzato con
	// --org-resource-id in run successivi senza ricreare la org.
	if input.SeedOrgName != "" {
		seedOrg, err := api.CreateSubOrganization(ctx, organization.ID, api.CreateOrganizationInDTO{
			Name: input.SeedOrgName,
		})
		if err != nil {
			return err
		}
		if err = api.SetAllLimitToUnlimited(ctx, seedOrg.ResourceID, time.Duration(input.RequestDelayMs)*time.Millisecond); err != nil {
			return err
		}
		utils.PrintlnGreen(fmt.Sprintf("Seed org created: %q  resource_id: %s", seedOrg.Name, seedOrg.ResourceID))

		// Sostituiamo "organization" con la nuova org contenitore:
		// da qui in poi device e sub-org vengono creati al suo interno.
		organization = &dto.OrganizationOutDTO{
			ID:         seedOrg.ID,
			ResourceID: seedOrg.ResourceID,
			Name:       seedOrg.Name,
		}
	}

	err = createDeviceGroups(ctx, input, organization)
	if err != nil {
		return err
	}

	err = createModels(ctx, input, organization)
	if err != nil {
		return err
	}

	err = createServiceAccounts(ctx, input, organization)
	if err != nil {
		return err
	}

	// L'albero viene creato PRIMA di device e utenti, così possiamo
	// raccogliere i riferimenti alle sub-org e distribuire su di esse.
	subOrgs, err := createOrganizations(ctx, input, organization.ID)
	if err != nil {
		return err
	}

	// I device vengono distribuiti sulle sub-org create (se presenti),
	// altrimenti finiscono tutti nella org radice come in precedenza.
	err = createDevices(ctx, input, organization, subOrgs)
	if err != nil {
		return err
	}

	// Gli utenti vengono creati nella root org (che ha il realm Keycloak),
	// non nelle sub-org che non hanno hostname configurato.
	err = createUsers(ctx, input, rootOrg, subOrgs)
	if err != nil {
		return err
	}

	utils.PrintlnBlue(fmt.Sprintf("Total elapsed: %s", time.Since(totalStart).Round(time.Second)))

	return nil
}

func createServiceAccounts(ctx context.Context, input *dto.ExecuteInDTO, organization *dto.OrganizationOutDTO) error {
	if input.ServiceAccountCount == 0 {
		return nil
	}

	for i := int64(0); i < input.ServiceAccountCount; i++ {
		_, err := api.CreateServiceAccount(ctx, organization.ID, utils.RandomName())
		if err != nil {
			return err
		}
	}

	utils.PrintlnGreen(fmt.Sprintf("Service accounts created: %d", input.ServiceAccountCount))

	return nil
}

// createDevices crea i device e li distribuisce round-robin sulle sub-org
// (se ne sono state create). Se subOrgs è vuoto, il comportamento è identico
// a prima: tutti i device vanno nella org radice.
func createDevices(ctx context.Context, input *dto.ExecuteInDTO, organization *dto.OrganizationOutDTO, subOrgs []api.CreateOrganizationOutDTO) error {
	if input.DeviceCount == 0 {
		return nil
	}

	devStart := time.Now()
	logEvery := input.DeviceCount / 10
	if logEvery < 1 {
		logEvery = 1
	}
	created := int64(0)

	for i := int64(0); i < input.DeviceCount; i++ {
		// Seleziona l'org target: round-robin sulle sub-org, oppure root org.
		var targetResourceID string
		if len(subOrgs) > 0 {
			targetResourceID = subOrgs[int(i)%len(subOrgs)].ResourceID
		} else {
			targetResourceID = organization.ResourceID
		}

		// Throttle: Elasticsearch (mappings) è lento; senza delay si satura.
		time.Sleep(time.Duration(input.RequestDelayMs) * time.Millisecond)
		_, err := api.CreateDevice(ctx, targetResourceID, utils.RandomName())
		if err != nil {
			// NON ritentiamo: CreateDevice è multi-step (crea device → mappings → licenza).
			// Se mappings va in timeout il device è già scritto nel DB; un retry usa lo stesso
			// stato parziale e causa "400 Error activating device license".
			log.Error().Err(err).Str("orgResourceId", targetResourceID).Msg("Error creating device")
		} else {
			created++
		}

		done := i + 1
		if done%logEvery == 0 || done == input.DeviceCount {
			elapsed := time.Since(devStart)
			rate := float64(done) / elapsed.Seconds()
			utils.PrintlnBlue(fmt.Sprintf("  devices: %d / %d (%.1f dev/s, ok: %d)", done, input.DeviceCount, rate, created))
		}
	}

	utils.PrintlnGreen(fmt.Sprintf("Devices created: %d / %d", created, input.DeviceCount))

	return nil
}

func createDeviceGroups(ctx context.Context, input *dto.ExecuteInDTO, organization *dto.OrganizationOutDTO) error {
	if input.DeviceGroupCount == 0 {
		return nil
	}

	for i := int64(0); i < input.DeviceGroupCount; i++ {
		err := api.CreateDeviceGroup(ctx, organization.ID, api.CreateDeviceGroupInDTO{
			Name: utils.RandomName(),
		})
		if err != nil {
			return err
		}
	}

	utils.PrintlnGreen(fmt.Sprintf("Device groups created: %d", input.DeviceGroupCount))

	return nil
}

func createModels(ctx context.Context, input *dto.ExecuteInDTO, organization *dto.OrganizationOutDTO) error {
	if input.ModelCount == 0 {
		return nil
	}

	for i := int64(0); i < input.ModelCount; i++ {
		output, err := api.CreateRandomModel(ctx, organization.ResourceID)
		if err != nil {
			return err
		}

		log.Debug().Str("model.id", output.Id).Msg("Model created")
	}

	utils.PrintlnGreen(fmt.Sprintf("Models created: %d", input.ModelCount))

	return nil
}

// createUsers crea gli utenti e li distribuisce round-robin sulle sub-org
// (se ne sono state create). Se subOrgs è vuoto, tutti gli utenti vanno
// nella org radice.
func createUsers(ctx context.Context, input *dto.ExecuteInDTO, organization *dto.OrganizationOutDTO, subOrgs []api.CreateOrganizationOutDTO) error {
	if input.UserCount == 0 {
		return nil
	}

	// Il dominio email deve corrispondere al realm dell'org (es. "@exor"):
	// le sub-org senza hostname configurato rifiutano domini esterni.
	realm, _ := ctx.Value(utils.UserRealm).(string)

	userStart := time.Now()
	logEvery := input.UserCount / 10
	if logEvery < 1 {
		logEvery = 1
	}

	created := int64(0)
	for i := int64(0); i < input.UserCount; i++ {
		// Gli utenti vengono sempre creati nella root org: solo lei ha il realm
		// Keycloak configurato. Le sub-org non hanno hostname e rifiutano la
		// creazione utenti con errore 15004.
		targetOrgID := organization.ID

		name := utils.RandomName()
		time.Sleep(time.Duration(input.RequestDelayMs) * time.Millisecond)
		_, err := api.CreateUser(ctx, targetOrgID, api.CreateUserInDTO{
			Email:                    name + "@" + realm,
			Username:                 name,
			EmailVerified:            true,
			TemporaryPassword:        true,
			ServiceAccount:           false,
			PasswordChangeInvitation: false,
			GroupPoliciesEnabled:     false,
			MemberOf:                 []int64{},
			RegularUserOf:            []int64{},
			AdminUserOf:              []int64{},
		})
		if err != nil {
			// Logga e continua: un errore su una singola org non deve bloccare
			// la creazione degli utenti rimanenti (stesso comportamento della creazione org).
			log.Error().Err(err).Int64("orgId", targetOrgID).Msg("Error creating user")
		} else {
			created++
		}

		done := i + 1
		if done%logEvery == 0 || done == input.UserCount {
			elapsed := time.Since(userStart)
			rate := float64(done) / elapsed.Seconds()
			utils.PrintlnBlue(fmt.Sprintf("  users: %d / %d (%.1f user/s, ok: %d)", done, input.UserCount, rate, created))
		}
	}

	// Mostra il conteggio reale, così è visibile se ci sono stati errori parziali.
	utils.PrintlnGreen(fmt.Sprintf("Users created: %d / %d", created, input.UserCount))

	return nil
}

func calculateNumberOfOrganizations(input *dto.ExecuteInDTO) int64 {
	sum := int64(0)

	for i := int64(0); i <= input.OrganizationTreeDepth; i++ {
		sum += utils.PowInt64(input.OrganizationCount, i)
	}

	return sum
}

// createOrganizations costruisce l'albero di sub-org e restituisce i riferimenti
// a tutte le org create, in modo che i caller possano distribuirvi device e utenti.
func createOrganizations(ctx context.Context, input *dto.ExecuteInDTO, organizationId int64) ([]api.CreateOrganizationOutDTO, error) {
	if input.OrganizationCount == 0 {
		return nil, nil
	}

	totalOrganizations := calculateNumberOfOrganizations(input)
	utils.PrintlnBlue(fmt.Sprintf("Creating organizations: %d children × depth %d = %d theoretical (workers: %d)", input.OrganizationCount, input.OrganizationTreeDepth, totalOrganizations, input.OrgWorkers))

	createdOrgs, err := createSubOrgMultipleWorkers(ctx, input, organizationId)
	if err != nil {
		return nil, err
	}

	// Stampa il conteggio reale (non quello teorico) così è subito evidente
	// quante org sono state effettivamente create in caso di errori parziali.
	utils.PrintlnGreen(fmt.Sprintf("Organizations created: %d / %d", len(createdOrgs), totalOrganizations))

	return createdOrgs, nil
}

type WorkerDTO struct {
	depth  int64
	parent int64
}

// createSubOrgMultipleWorkers costruisce ricorsivamente l'albero di sub-org
// usando un pool di goroutine. Rispetto alla versione precedente, ora raccoglie
// e restituisce i riferimenti a tutte le org create (thread-safe via mutex),
// così il caller può distribuirvi device e utenti.
func createSubOrgMultipleWorkers(ctx context.Context, input *dto.ExecuteInDTO, organizationId int64) ([]api.CreateOrganizationOutDTO, error) {
	depth := input.OrganizationTreeDepth
	leafs := input.OrganizationCount
	// Numero di goroutine concorrenti: configurabile via --org-workers.
	// Un valore basso (default 3) protegge il connection pool di Postgres
	// su ambienti con risorse limitate (es. minikube).
	numOfDepthWorkers := int(input.OrgWorkers)
	var c = make(chan WorkerDTO, utils.PowInt64(input.OrganizationCount, input.OrganizationTreeDepth))
	wg := sync.WaitGroup{}

	// Slice condivisa tra le goroutine: ogni worker vi appende le org create.
	var createdOrgs []api.CreateOrganizationOutDTO
	var mu sync.Mutex

	orgStart := time.Now()
	totalOrgs := calculateNumberOfOrganizations(input)
	logEvery := totalOrgs / 10
	if logEvery < 1 {
		logEvery = 1
	}

	wg.Add(1)
	c <- WorkerDTO{depth: depth, parent: organizationId}

	for i := 0; i < numOfDepthWorkers; i++ {
		go func(workId int) {
			for dto := range c {
				if dto.depth == 0 {
					wg.Done()
					continue
				}

				for i := int64(0); i < leafs; i++ {
					subOrganization, err := createSubOrg(ctx, dto.parent, time.Duration(input.RequestDelayMs)*time.Millisecond)
					if err != nil {
						log.Error().Err(err).Int64("organziationId", organizationId).Msg("Error creating sub organization")
						continue
					}

					// Registra la sub-org creata nella lista condivisa.
					mu.Lock()
					createdOrgs = append(createdOrgs, *subOrganization)
					n := int64(len(createdOrgs))
					mu.Unlock()

					if n%logEvery == 0 || n == totalOrgs {
						elapsed := time.Since(orgStart)
						rate := float64(n) / elapsed.Seconds()
						eta := ""
						if rate > 0 && n < totalOrgs {
							remaining := time.Duration(float64(totalOrgs-n)/rate) * time.Second
							eta = " ETA: ~" + remaining.Round(time.Second).String()
						}
						utils.PrintlnBlue(fmt.Sprintf("  orgs: %d / %d (%.1f org/s%s)", n, totalOrgs, rate, eta))
					}

					log.Debug().Msgf("Worker %d, parent %d, depth %d\n", workId, dto.parent, depth-dto.depth+1)
					if dto.depth > 0 {
						wg.Add(1)
						c <- WorkerDTO{depth: dto.depth - 1, parent: subOrganization.ID}
					}
				}
				wg.Done()
			}
		}(i)
	}

	// Aspetta che tutte le work-unit siano completate, poi chiude il canale
	// per far uscire le goroutine dal loro loop range.
	wg.Wait()
	close(c)

	return createdOrgs, nil
}

// isRetryable distingue errori transitori da errori permanenti.
// 5xx = problemi server; 429 = rate limit (il server chiede di rallentare);
// timeout/connection refused = problemi di rete.
// I 4xx diversi da 429 (es. 400, 403, 404) sono errori permanenti: non ha
// senso riprovare perché il server rifiuterà la stessa richiesta di nuovo.
func isRetryable(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "Status:5") ||
		strings.Contains(msg, "Status:429") ||
		strings.Contains(msg, "deadline exceeded") ||
		strings.Contains(msg, "connection refused")
}

// withRetry esegue fn fino a 3 volte con backoff esponenziale (1s, 2s, 4s),
// ma solo se l'errore è classificato come transitorio da isRetryable.
func withRetry(fn func() error) error {
	const maxAttempts = 3
	delay := time.Second

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
			if attempt < maxAttempts && isRetryable(err) {
				log.Warn().Err(err).Int("attempt", attempt).Dur("retry_in", delay).Msg("errore transitorio, riprovo")
				time.Sleep(delay)
				delay *= 2
				continue
			}
			return err
		}
	}
	return fmt.Errorf("max tentativi superati: %w", lastErr)
}

// createSubOrg crea una sub-org e ne imposta i limiti a illimitato.
// delay è il tempo minimo di attesa prima di ogni chiamata HTTP: serve a
// evitare di saturare il rate limiter dell'API anche con pochi worker.
func createSubOrg(ctx context.Context, organizationId int64, delay time.Duration) (*api.CreateOrganizationOutDTO, error) {
	// Step 1: crea la sub-org con retry su errori transitori.
	// Il delay precede ogni tentativo: anche i retry rispettano il throttle.
	// Ogni tentativo genera un nome nuovo per evitare duplicati nel caso in cui
	// la richiesta HTTP sia arrivata al server ma la risposta sia andata persa.
	var subOrganization *api.CreateOrganizationOutDTO
	if err := withRetry(func() error {
		time.Sleep(delay)
		var e error
		subOrganization, e = api.CreateSubOrganization(ctx, organizationId, api.CreateOrganizationInDTO{
			Name: utils.RandomName(),
		})
		return e
	}); err != nil {
		return nil, err
	}

	// Step 2: imposta i limiti con retry separato.
	// La org esiste già: non c'è rischio di duplicati in caso di nuovo tentativo.
	if err := withRetry(func() error {
		time.Sleep(delay)
		return api.SetAllLimitToUnlimited(ctx, subOrganization.ResourceID, delay)
	}); err != nil {
		return nil, err
	}

	return subOrganization, nil
}
