package utils

type CtxKey string

const OriginKey CtxKey = "origin"
const DomainKey CtxKey = "domain"
const DeviceCount CtxKey = "device-count"
const EachDeviceHasMapping CtxKey = "each-device-has-mapping"
const DeviceGroupCount CtxKey = "device-group-count"
const ModelCount CtxKey = "model-count"
const ServiceAccountCount CtxKey = "service-account-count"
const OrganizationCount CtxKey = "organization-count"
const OrganizationTreeDepth CtxKey = "organization-tree-depth"
const KeycloakOrigin CtxKey = "keycloak-host"
const KeycloakMasterClientId CtxKey = "keycloak-master-client-id"
const KeycloakMasterClientSecret CtxKey = "keycloak-master-client-secret"
const AdminUserKey CtxKey = "admin-user"
const UserRealm CtxKey = "user-realm"
const LicenseHostKey CtxKey = "license-endpoint"
const LicenseManagerClientId CtxKey = "license-manager-client-id"
const LicenseManagerClientSecret CtxKey = "license-manager-client-secret"
