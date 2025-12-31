import { User } from '../../types/user'

interface UserAvatarProps {
  user: User
}

export function UserAvatar({ user }: UserAvatarProps) {
  return (
    <div className="flex items-center gap-3">
      <img
        src={user.picture}
        alt={user.name}
        className="w-8 h-8 rounded-full"
      />
      <span className="text-sm text-gray-300">{user.name}</span>
      <a
        href="/auth/logout"
        className="text-sm text-gray-400 hover:text-white"
      >
        Sign out
      </a>
    </div>
  )
}
