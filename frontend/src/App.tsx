import React from 'react';
import { BrowserRouter as Router, Route, Routes, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from 'react-query';
import { AuthProvider } from './contexts/AuthContext';
import { useAuth } from './contexts/UseAuth';
import Login from './routes/Login';
import Player from './routes/Player';
import PlaylistMaker from './routes/PlaylistMaker';
import MenuBar from './components/MenuBar';

const queryClient = new QueryClient();

interface ProtectedRouteProps {
  children: React.ReactNode;
}

const ProtectedRoute: React.FC<ProtectedRouteProps> = ({ children }) => {
  const { isAuthenticated } = useAuth();
  return isAuthenticated ? <>{children}</> : <Navigate to="/login" replace />;
};

const AuthenticatedLayout: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  return (
    <>
      <MenuBar />
      {children}
    </>
  );
};

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <Router>
          <Routes>
            <Route path="/login" element={<Login />} />
            <Route path="/player" element={
              <ProtectedRoute>
                <AuthenticatedLayout>
                  <Player />
                </AuthenticatedLayout>
              </ProtectedRoute>
            } />
            <Route path="/playlist-maker" element={
              <ProtectedRoute>
                <AuthenticatedLayout>
                  <PlaylistMaker />
                </AuthenticatedLayout>
              </ProtectedRoute>
            } />
            <Route path="/" element={<Navigate replace to="/login" />} />
          </Routes>
        </Router>
      </AuthProvider>
    </QueryClientProvider>
  );
}

export default App;