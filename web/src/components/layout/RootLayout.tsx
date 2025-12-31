import { Outlet } from 'react-router-dom'
import { Navigation } from './Navigation'

export function RootLayout() {
  return (
    <div className="min-h-screen bg-gray-900 text-white">
      <Navigation />
      <main className="max-w-4xl mx-auto px-4 py-8">
        <Outlet />
      </main>
    </div>
  )
}
