import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './contexts/AuthContext';
import ProtectedRoute from './components/ProtectedRoute';
import Layout from './components/Layout';

// Auth Pages
import Login from './pages/Login';
import Setup2FA from './pages/Setup2FA';
import Verify2FA from './pages/Verify2FA';
import VerifyOTP from './pages/VerifyOTP';
import TestPage from './pages/TestPage';

// Dashboard & Management Pages
import Dashboard from './pages/Dashboard';
import Services from './pages/Services';
import Workflows from './pages/Workflows';
import MinIO from './pages/MinIO';
import Queues from './pages/Queues';
import WorkflowEditor from './components/WorkflowEditor';

import './App.css';

function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          {/* Public Routes */}
          <Route path="/login" element={<Login />} />
          <Route path="/setup-2fa" element={<Setup2FA />} />
          <Route path="/verify-2fa" element={<Verify2FA />} />
          <Route path="/verify-otp" element={<VerifyOTP />} />
          <Route path="/test" element={<TestPage />} />

          {/* Workflow Editor Routes (Full Screen - No Layout) */}
          <Route path="/workflows/new" element={
            <ProtectedRoute>
              <WorkflowEditor />
            </ProtectedRoute>
          } />
          <Route path="/workflows/edit/:id" element={
            <ProtectedRoute>
              <WorkflowEditor />
            </ProtectedRoute>
          } />

          {/* Protected Routes */}
          <Route
            path="/"
            element={
              <ProtectedRoute>
                <Layout />
              </ProtectedRoute>
            }
          >
            <Route index element={<Navigate to="/dashboard" replace />} />
            <Route path="dashboard" element={<Dashboard />} />
            <Route path="services" element={<Services />} />
            <Route path="workflows" element={<Workflows />} />
            <Route path="minio" element={<MinIO />} />
            <Route path="queues" element={<Queues />} />
          </Route>

          {/* Catch all */}
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}

export default App;
