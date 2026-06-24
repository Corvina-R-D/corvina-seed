package dto

type ExecuteInDTO struct {
	DeviceCount           int64
	DeviceGroupCount      int64
	ModelCount            int64
	ServiceAccountCount   int64
	OrganizationCount     int64
	OrganizationTreeDepth int64
	UserCount             int64  // utenti da distribuire sull'albero di org
	SeedOrgName           string // nome della org contenitore da creare come primo passo
	OrgWorkers            int64  // goroutine concorrenti per la creazione dell'albero (default 3)
	RequestDelayMs        int64  // pausa minima in ms tra ogni chiamata API (default 200, evita il rate limiter)
}
