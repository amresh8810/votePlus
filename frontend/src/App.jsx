import { Routes, Route, Navigate } from 'react-router-dom';
import { useAuth } from './context/AuthContext';
import { Navbar } from './components/Navbar';
import { ProtectedRoute } from './components/ProtectedRoute';
import { Loading } from './components/Loading';

// Pages
import { Login } from './pages/Login';
import { Signup } from './pages/Signup';
import { VerifyEmail } from './pages/VerifyEmail';
import { ForgotPassword } from './pages/ForgotPassword';
import { ResetPassword } from './pages/ResetPassword';
import { Dashboard } from './pages/Dashboard';
import { CreatePoll } from './pages/CreatePoll';
import { ManagePoll } from './pages/ManagePoll';
import { PublicPoll } from './pages/PublicPoll';
import { Account } from './pages/Account';
import { UiProvider } from './context/UiContext';

function HomeRedirect() {
  const { isAuthenticated, loading } = useAuth();
  if (loading) return <Loading message="Loading VotePulse..." fullPage />;
  return <Navigate to={isAuthenticated ? '/dashboard' : '/login'} replace />;
}

function GuestOnlyRoute({ children }) {
  const { isAuthenticated, loading } = useAuth();
  if (loading) return <Loading message="Loading VotePulse..." fullPage />;
  if (isAuthenticated) return <Navigate to="/dashboard" replace />;
  return children;
}

function NotFound() {
  return (
    <div className="page-layout text-center py-16">
      <h1 className="text-4xl font-extrabold text-indigo-400 mb-3">404</h1>
      <h2 className="text-xl font-bold mb-2">Page Not Found</h2>
      <p className="text-slate-400 mb-6">The page you are looking for does not exist or has been moved.</p>
      <a href="/" className="btn btn-primary btn-md">
        Return Home
      </a>
    </div>
  );
}

function App() {
  return (
    <UiProvider>
    <div className="app-container">
      <Navbar />
      <main className="main-content">
        <Routes>
          <Route path="/" element={<HomeRedirect />} />

          {/* Public Auth Routes */}
          <Route path="/login" element={<GuestOnlyRoute><Login /></GuestOnlyRoute>} />
          <Route path="/signup" element={<GuestOnlyRoute><Signup /></GuestOnlyRoute>} />
          <Route path="/verify-email" element={<VerifyEmail />} />
          <Route path="/forgot-password" element={<ForgotPassword />} />
          <Route path="/reset-password" element={<ResetPassword />} />

          {/* Public Audience Voting Route */}
          <Route path="/p/:id" element={<PublicPoll />} />

          {/* Protected Creator Routes */}
          <Route
            path="/dashboard"
            element={
              <ProtectedRoute>
                <Dashboard />
              </ProtectedRoute>
            }
          />
          <Route
            path="/polls/create"
            element={
              <ProtectedRoute>
                <CreatePoll />
              </ProtectedRoute>
            }
          />
          <Route
            path="/polls/:id"
            element={
              <ProtectedRoute>
                <ManagePoll />
              </ProtectedRoute>
            }
          />
          <Route
            path="/account"
            element={<ProtectedRoute><Account /></ProtectedRoute>}
          />

          {/* 404 Catch All */}
          <Route path="*" element={<NotFound />} />
        </Routes>
      </main>
    </div>
    </UiProvider>
  );
}

export default App;
