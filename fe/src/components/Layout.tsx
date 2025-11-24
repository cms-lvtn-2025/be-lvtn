import React from 'react';
import { Outlet } from 'react-router-dom';
import Sidebar from './Sidebar';
import { useAuth } from '../contexts/AuthContext';

const Layout: React.FC = () => {
  const { user, logout } = useAuth();

  return (
    <div className="layout">
      <Sidebar />
      <div className="main-content">
        <header className="header">
          <h1>Plagiarism Checker Admin</h1>
          <div className="user-info">
            <span>{user?.email}</span>
            <span className="role-badge">{user?.role}</span>
            <button onClick={logout} className="btn btn-sm btn-outline">
              Logout
            </button>
          </div>
        </header>
        <div className="content">
          <Outlet />
        </div>
      </div>
    </div>
  );
};

export default Layout;
