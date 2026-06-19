export interface UserStore {
  id: number
  name: string
  avatar: string
  avatarMin: string
  moemoepoint: number
  role: number
  // Holds the creator role (orthogonal to the numeric role). Refreshed per page
  // from /user/status; gates the avatar-menu "创作者申请" entry.
  isCreator: boolean
  isCheckIn: boolean
  dailyToolsetUploadBytes: number
}
