interface OverallStatusProps {
  status: string
}

export function OverallStatus({ status }: OverallStatusProps) {
  const isOk = status === 'ok'

  return (
    <div
      className={`rounded-lg p-4 ${
        isOk
          ? 'bg-green-900/50 border border-green-500'
          : 'bg-yellow-900/50 border border-yellow-500'
      }`}
    >
      <p className="text-sm text-gray-400">Overall Status</p>
      <p
        className={`text-xl font-semibold ${
          isOk ? 'text-green-400' : 'text-yellow-400'
        }`}
      >
        {status.toUpperCase()}
      </p>
    </div>
  )
}
