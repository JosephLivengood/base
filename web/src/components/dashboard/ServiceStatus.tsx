interface ServiceStatusProps {
  name: string
  status: string
}

export function ServiceStatus({ name, status }: ServiceStatusProps) {
  const isHealthy = status === 'healthy'

  return (
    <div className="p-4 flex items-center justify-between">
      <span className="font-medium">{name}</span>
      <div className="flex items-center gap-2">
        <span
          className={`w-2 h-2 rounded-full ${
            isHealthy ? 'bg-green-500' : 'bg-red-500'
          }`}
        />
        <span
          className={`text-sm ${isHealthy ? 'text-green-400' : 'text-red-400'}`}
        >
          {isHealthy ? 'healthy' : status}
        </span>
      </div>
    </div>
  )
}
