package seed

import (
	"context"
	"corvina/corvina-seed/src/seed/api"
	"corvina/corvina-seed/src/seed/dto"
	"corvina/corvina-seed/src/utils"
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"
)

func Execute(ctx context.Context, input *dto.ExecuteInDTO) error {

	organization, err := api.GetOrganizationMine(ctx)
	if err != nil {
		return err
	}

	log.Info().Interface("organization", organization).Msg("Organization retrieved")

	// Se è stato fornito --seed-org-name, creiamo una sub-org con quel nome
	// come "contenitore" di tutto ciò che segue (albero, utenti, device).
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
		if err = api.SetAllLimitToUnlimited(ctx, seedOrg.ResourceID); err != nil {
			return err
		}
		utils.PrintlnGreen(fmt.Sprintf("Seed org created: %q  resource_id: %s", seedOrg.Name, seedOrg.ResourceID))

		// Sostituiamo "organization" con la nuova org contenitore:
		// da qui in poi device, utenti e sub-org vengono creati al suo interno.
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

	// Gli utenti vengono distribuiti allo stesso modo.
	err = createUsers(ctx, input, organization, subOrgs)
	if err != nil {
		return err
	}

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

	for i := int64(0); i < input.DeviceCount; i++ {
		// Seleziona l'org target: round-robin sulle sub-org, oppure root org.
		var targetResourceID string
		if len(subOrgs) > 0 {
			targetResourceID = subOrgs[int(i)%len(subOrgs)].ResourceID
		} else {
			targetResourceID = organization.ResourceID
		}

		_, err := api.CreateDevice(ctx, targetResourceID, utils.RandomName())
		if err != nil {
			return err
		}
	}

	utils.PrintlnGreen(fmt.Sprintf("Devices created: %d", input.DeviceCount))

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

	for i := int64(0); i < input.UserCount; i++ {
		// Seleziona l'org target: round-robin sulle sub-org, oppure root org.
		var targetOrgID int64
		if len(subOrgs) > 0 {
			targetOrgID = subOrgs[int(i)%len(subOrgs)].ID
		} else {
			targetOrgID = organization.ID
		}

		name := utils.RandomName()
		_, err := api.CreateUser(ctx, targetOrgID, api.CreateUserInDTO{
			Email:                    name + "@seed.corvina.io",
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
			return err
		}
	}

	utils.PrintlnGreen(fmt.Sprintf("Users created: %d", input.UserCount))

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
	utils.PrintlnBlue(fmt.Sprintf("Creating organizations: %d with depth %d = %d", input.OrganizationCount, input.OrganizationTreeDepth, totalOrganizations))

	createdOrgs, err := createSubOrgMultipleWorkers(ctx, input, organizationId)
	if err != nil {
		return nil, err
	}

	utils.PrintlnGreen(fmt.Sprintf("Organizations created: %d", totalOrganizations))

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
	numOfDepthWorkers := 20
	var c = make(chan WorkerDTO, utils.PowInt64(input.OrganizationCount, input.OrganizationTreeDepth))
	wg := sync.WaitGroup{}

	// Slice condivisa tra le goroutine: ogni worker vi appende le org create.
	var createdOrgs []api.CreateOrganizationOutDTO
	var mu sync.Mutex

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
					subOrganization, err := createSubOrg(ctx, dto.parent)
					if err != nil {
						log.Error().Err(err).Int64("organziationId", organizationId).Msg("Error creating sub organization")
						continue
					}

					// Registra la sub-org creata nella lista condivisa.
					mu.Lock()
					createdOrgs = append(createdOrgs, *subOrganization)
					mu.Unlock()

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

func createSubOrg(ctx context.Context, organizationId int64) (*api.CreateOrganizationOutDTO, error) {
	createSubOrgInput := api.CreateOrganizationInDTO{
		Name: utils.RandomName(),
	}
	subOrganization, err := api.CreateSubOrganization(ctx, organizationId, createSubOrgInput)
	if err != nil {
		return nil, err
	}
	err = api.SetAllLimitToUnlimited(ctx, subOrganization.ResourceID)
	if err != nil {
		return nil, err
	}
	return subOrganization, err
}
