import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { authApi } from '../api/auth';

const Login: React.FC = () => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      const response = await authApi.login(email, password);

      if (response.success) {
        // Store temp token
        localStorage.setItem('tempToken', response.data.tempToken);
        localStorage.setItem('userEmail', response.data.email);

        // Check if 2FA is enabled
        if (response.data.twoFactorEnabled) {
          navigate('/verify-2fa');
        } else {
          // First time login, setup 2FA
          navigate('/setup-2fa');
        }
      }
    } catch (err: any) {
      setError(err.response?.data?.message || 'Login failed. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="auth-container">
      <div className="auth-card">
        <h1>Plagiarism Checker Admin</h1>
        <h2>Login</h2>

        {error && <div className="alert alert-error">{error}</div>}

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="email">Email</label>
            <input
              type="email"
              id="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="lyvinhthai321@gmail.com"
              required
              autoFocus
            />
          </div>

          <div className="form-group">
            <label htmlFor="password">Password</label>
            <input
              type="password"
              id="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Enter your password"
              required
            />
          </div>

          <button type="submit" className="btn btn-primary btn-block" disabled={loading}>
            {loading ? 'Logging in...' : 'Login'}
          </button>
        </form>

        <div className="auth-footer">
          <p>Default credentials:</p>
          <p>
            <strong>Email:</strong> lyvinhthai321@gmail.com
          </p>
          <p>
            <strong>Password:</strong> Admin@123
          </p>
        </div>
      </div>
    </div>
  );
};

export default Login;
