import { createBrowserRouter } from 'react-router-dom'
import { RootLayout } from '../components/layout/RootLayout'
import { ProtectedRoute } from '../components/auth/ProtectedRoute'
import { Dashboard } from '../pages/Dashboard'
import { Login } from '../pages/Login'
import { Profile } from '../pages/Profile'
import { Organizations } from '../pages/Organizations'
import { OrgSettings } from '../pages/OrgSettings'
import { Invitations } from '../pages/Invitations'
import { NotFound } from '../pages/NotFound'

export const router = createBrowserRouter([
  {
    path: '/',
    element: <RootLayout />,
    children: [
      {
        index: true,
        element: <Dashboard />,
      },
      {
        element: <ProtectedRoute />,
        children: [
          {
            path: 'profile',
            element: <Profile />,
          },
          {
            path: 'organizations',
            element: <Organizations />,
          },
          {
            path: 'organizations/:orgId/settings',
            element: <OrgSettings />,
          },
          {
            path: 'invitations',
            element: <Invitations />,
          },
        ],
      },
    ],
  },
  {
    path: '/login',
    element: <Login />,
  },
  {
    path: '*',
    element: <NotFound />,
  },
])
