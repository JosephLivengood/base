import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { RouterProvider } from 'react-router-dom'
import './index.css'
import { AuthProvider } from './hooks/useAuth'
import { OrganizationProvider } from './hooks/useOrganization'
import { router } from './router'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <AuthProvider>
      <OrganizationProvider>
        <RouterProvider router={router} />
      </OrganizationProvider>
    </AuthProvider>
  </StrictMode>,
)
