import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { authApi } from '../api/auth';
import { useAuth } from '../contexts/AuthContext';

const Setup2FA: React.FC = () => {
  const [qrCodeUrl, setQrCodeUrl] = useState('');
  const [secret, setSecret] = useState('');
  const [token, setToken] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [setupLoading, setSetupLoading] = useState(true);
  const navigate = useNavigate();
  const { login } = useAuth();

  useEffect(() => {
    const setup = async () => {
      const tempToken = localStorage.getItem('tempToken');
      if (!tempToken) {
        navigate('/login');
        return;
      }

      try {
        const response = await authApi.setup2FA(tempToken);
        console.log('Setup 2FA response:', response);
        if (response.success) {
          setQrCodeUrl(response.data.qrCodeUrl);
          setSecret(response.data.secret);
        }
      } catch (err: any) {
        console.error('Setup 2FA error:', err);
        setError(err.response?.data?.message || 'Failed to setup 2FA');
      } finally {
        setSetupLoading(false);
      }
    };

    setup();
  }, [navigate]);

  const handleVerify = async (e: React.FormEvent) => {
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

  if (setupLoading) {
    return (
      <div className="auth-container">
        <div className="auth-card">
          <div className="loading-container">
            <div className="spinner"></div>
            <p>Setting up 2FA...</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="auth-container">
      <div className="auth-card">
        <h2>Setup Two-Factor Authentication</h2>

        {error && <div className="alert alert-error">{error}</div>}

        <div className="setup-2fa-content">
          <div className="qr-section">
            <p>Scan this QR code with Google Authenticator app:</p>
            {qrCodeUrl && (
              <div className="qr-code">
                <img src={qrCodeUrl} alt="QR Code" style={{ width: '200px', height: '200px' }} />
              </div>
            )}
          </div>

          <div className="manual-entry">
            <p>Or enter this code manually:</p>
            <code className="secret-code">{secret}</code>
          </div>

          <form onSubmit={handleVerify}>
            <div className="form-group">
              <label htmlFor="token">Enter 6-digit code from app</label>
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
              {loading ? 'Verifying...' : 'Verify & Continue'}
            </button>
          </form>

          <div className="auth-divider">OR</div>

          <button onClick={handleUseEmailOTP} className="btn btn-outline btn-block">
            Use Email OTP Instead
          </button>
        </div>
      </div>
    </div>
  );
};

export default Setup2FA;
