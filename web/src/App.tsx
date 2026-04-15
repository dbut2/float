import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import BucketDetail from './pages/BucketDetail'
import Settings from './pages/Settings'
import Onboarding from './pages/Onboarding'
import Health from './pages/Health'

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="onboarding" element={<Onboarding />} />
        <Route element={<Layout />}>
          <Route index element={<Dashboard />} />
          <Route path="buckets/:id" element={<BucketDetail />} />
          <Route path="health" element={<Health />} />
<Route path="settings" element={<Settings />} />
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  )
}
