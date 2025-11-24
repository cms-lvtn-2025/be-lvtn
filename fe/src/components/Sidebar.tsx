import React, { useState, useEffect } from 'react';
import { NavLink } from 'react-router-dom';

const Sidebar: React.FC = () => {
  const [isCollapsed, setIsCollapsed] = useState(() => {
    const saved = localStorage.getItem('sidebarCollapsed');
    return saved === 'true';
  });

  useEffect(() => {
    localStorage.setItem('sidebarCollapsed', String(isCollapsed));
    document.documentElement.setAttribute('data-sidebar-collapsed', String(isCollapsed));
  }, [isCollapsed]);

  const toggleSidebar = () => {
    setIsCollapsed(!isCollapsed);
  };

  return (
    <aside className={`sidebar ${isCollapsed ? 'collapsed' : ''}`}>
      <div className="sidebar-header">
        <h2 className="sidebar-title">{!isCollapsed && 'Admin Panel'}</h2>
        <button
          className="sidebar-toggle"
          onClick={toggleSidebar}
          title={isCollapsed ? 'Expand sidebar' : 'Collapse sidebar'}
        >
          {isCollapsed ? '☰' : '‹'}
        </button>
      </div>
      <nav className="sidebar-nav">
        <NavLink to="/dashboard" className={({ isActive }) => (isActive ? 'nav-item active' : 'nav-item')} title="Dashboard">
          <span className="nav-icon">📊</span>
          {!isCollapsed && <span className="nav-text">Dashboard</span>}
        </NavLink>
        <NavLink to="/services" className={({ isActive }) => (isActive ? 'nav-item active' : 'nav-item')} title="Services">
          <span className="nav-icon">⚙️</span>
          {!isCollapsed && <span className="nav-text">Services</span>}
        </NavLink>
        <NavLink to="/workflows" className={({ isActive }) => (isActive ? 'nav-item active' : 'nav-item')} title="Workflows">
          <span className="nav-icon">🔄</span>
          {!isCollapsed && <span className="nav-text">Workflows</span>}
        </NavLink>
        <NavLink to="/minio" className={({ isActive }) => (isActive ? 'nav-item active' : 'nav-item')} title="MinIO Storage">
          <span className="nav-icon">💾</span>
          {!isCollapsed && <span className="nav-text">MinIO Storage</span>}
        </NavLink>
        <NavLink to="/queues" className={({ isActive }) => (isActive ? 'nav-item active' : 'nav-item')} title="Queues">
          <span className="nav-icon">📥</span>
          {!isCollapsed && <span className="nav-text">Queues</span>}
        </NavLink>
      </nav>
    </aside>
  );
};

export default Sidebar;
