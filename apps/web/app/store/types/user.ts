export interface UserStore {
  id: number
  // OAuth subject UUID — the stable account identity used as the account-switch
  // login_hint (the local integer id is not what OAuth keys on).
  sub: string
  name: string
  avatar: string
  avatarMin: string
  moemoepoint: number
  role: number
  // Raw OAuth role list (e.g. ["user","admin"]) — drives the account-switcher's
  // admin badge; `role` is the derived numeric.
  roles: string[]
  // Holds the creator role (orthogonal to the numeric role). Refreshed per page
  // from /user/status; gates the avatar-menu "创作者申请" entry.
  isCreator: boolean
  isCheckIn: boolean
  dailyToolsetUploadBytes: number
}
