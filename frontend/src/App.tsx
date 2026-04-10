import React from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import ProtectedRoute from './components/ProtectedRoute';
import Navbar from './components/Navbar';
import LoginPage from './pages/LoginPage';
import RegisterPage from './pages/RegisterPage';
import ProjectsPage from './pages/ProjectsPage';
import ProjectDetailPage from './pages/ProjectDetailPage';
import MyTasksPage from './pages/MyTasksPage';

const App: React.FC = () => {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Navbar />
        <main className="app-main">
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/register" element={<RegisterPage />} />
            <Route
              path="/projects"
              element={
                <ProtectedRoute>
                  <ProjectsPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/my-tasks"
              element={
                <ProtectedRoute>
                  <MyTasksPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/projects/:id"
              element={
                <ProtectedRoute>
                  <ProjectDetailPage />
                </ProtectedRoute>
              }
            />
            <Route path="/" element={<Navigate to="/projects" replace />} />
            <Route path="*" element={<Navigate to="/projects" replace />} />
          </Routes>
        </main>
      </AuthProvider>
    </BrowserRouter>
  );
};

export default App;
