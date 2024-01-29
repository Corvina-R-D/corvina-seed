package utils

type CtxKey string

const OriginKey CtxKey = "origin"
const DomainKey CtxKey = "domain"
const DeviceCount CtxKey = "device-count"
const DeviceGroupCount CtxKey = "device-group-count"
const ModelCount CtxKey = "model-count"
const KeycloakOrigin CtxKey = "keycloak-host"
const KeycloakMasterUser CtxKey = "keycloak-master-user"
const KeycloakMasterPass CtxKey = "keycloak-master-pass"
const AdminUserKey CtxKey = "admin-user"
const UserRealm CtxKey = "user-realm"
