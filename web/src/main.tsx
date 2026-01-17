import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { RouterProvider } from 'react-router-dom'
import HyperDX from '@hyperdx/browser'
import './index.css'
import { AuthProvider } from './hooks/useAuth'
import { OrganizationProvider } from './hooks/useOrganization'
import { router } from './router'

// Initialize HyperDX for browser observability
HyperDX.init({
  apiKey: 'local-dev', // Not validated in self-hosted
  service: 'base2-web',
  url: import.meta.env.VITE_HYPERDX_ENDPOINT || 'http://localhost:4318',
  tracePropagationTargets: [/localhost/i, /127\.0\.0\.1/i],
  consoleCapture: true,
  advancedNetworkCapture: true,
})

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <AuthProvider>
      <OrganizationProvider>
        <RouterProvider router={router} />
      </OrganizationProvider>
    </AuthProvider>
  </StrictMode>,
)
