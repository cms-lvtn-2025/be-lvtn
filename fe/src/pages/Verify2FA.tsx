import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { authApi } from '../api/auth';
import { useAuth } from '../contexts/AuthContext';

const Verify2FA: React.FC = () => {
  const [token, setToken] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const { login } = useAuth();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    const tempToken = localStorage.getItem('tempToken');
    if (!tempToken) {
      navigate('/login');
      return;
    }

    try {
      const response = await authApi.verify2FA(tempToken, token);

      if (response.success) {
        login(response.data.accessToken, response.data.user);
        localStorage.removeItem('tempToken');
        navigate('/dashboard');
      }
    } catch (err: any) {
      setError(err.response?.data?.message || 'Invalid 2FA token');
    } finally {
      setLoading(false);
    }
  };

  const handleUseEmailOTP = () => {
    navigate('/verify-otp');
  };

  return (
    <div className="auth-container">
      <div className="auth-card">
        <h2>Two-Factor Authentication</h2>
        <p>Enter the 6-digit code from your Google Authenticator app</p>

        {error && <div className="alert alert-error">{error}</div>}

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="token">Authentication Code</label>
            <input
              type="text"
              id="token"
              value={token}
              onChange={(e) => setToken(e.target.value.replace(/\D/g, '').slice(0, 6))}
              placeholder="123456"
              maxLength={6}
              required
              autoFocus
            />
          </div>

          <button type="submit" className="btn btn-primary btn-block" disabled={loading}>
            {loading ? 'Verifying...' : 'Verify'}
          </button>
        </form>

        <div className="auth-divider">OR</div>

        <button onClick={handleUseEmailOTP} className="btn btn-outline btn-block">
          Use Email OTP Instead
        </button>

        <div className="auth-footer">
          <a href="/login" className="text-link">
            Back to Login
          </a>
        </div>
      </div>
    </div>
  );
};

export default Verify2FA;
