import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { authApi } from '../api/auth';
import { useAuth } from '../contexts/AuthContext';

const VerifyOTP: React.FC = () => {
  const [otp, setOtp] = useState('');
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [loading, setLoading] = useState(false);
  const [sendingOTP, setSendingOTP] = useState(false);
  const [otpSent, setOtpSent] = useState(false);
  const navigate = useNavigate();
  const { login } = useAuth();

  useEffect(() => {
    // Auto send OTP on mount
    handleSendOTP();
  }, []);

  const handleSendOTP = async () => {
    const tempToken = localStorage.getItem('tempToken');
    if (!tempToken) {
      navigate('/login');
      return;
    }

    setSendingOTP(true);
    setError('');

    try {
      const response = await authApi.requestOTPEmail(tempToken);
      if (response.success) {
        setSuccess(`OTP sent to ${response.data?.email || 'your email'}`);
        setOtpSent(true);
      }
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to send OTP');
    } finally {
      setSendingOTP(false);
    }
  };

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
      const response = await authApi.verifyOTPEmail(tempToken, otp);

      if (response.success) {
        login(response.data.accessToken, response.data.user);
        localStorage.removeItem('tempToken');
        navigate('/dashboard');
      }
    } catch (err: any) {
      setError(err.response?.data?.message || 'Invalid or expired OTP');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="auth-container">
      <div className="auth-card">
        <h2>Email OTP Verification</h2>
        <p>Enter the 6-digit code sent to your email</p>

        {error && <div className="alert alert-error">{error}</div>}
        {success && <div className="alert alert-success">{success}</div>}

        {sendingOTP ? (
          <div className="loading-container">
            <div className="spinner"></div>
            <p>Sending OTP...</p>
          </div>
        ) : (
          <>
            <form onSubmit={handleSubmit}>
              <div className="form-group">
                <label htmlFor="otp">OTP Code</label>
                <input
                  type="text"
                  id="otp"
                  value={otp}
                  onChange={(e) => setOtp(e.target.value.replace(/\D/g, '').slice(0, 6))}
                  placeholder="123456"
                  maxLength={6}
                  required
                  autoFocus
                  disabled={!otpSent}
                />
              </div>

              <button type="submit" className="btn btn-primary btn-block" disabled={loading || !otpSent}>
                {loading ? 'Verifying...' : 'Verify OTP'}
              </button>
            </form>

            <div className="auth-footer">
              <button onClick={handleSendOTP} className="text-link" disabled={sendingOTP}>
                Resend OTP
              </button>
              <span> | </span>
              <a href="/verify-2fa" className="text-link">
                Use 2FA Instead
              </a>
              <span> | </span>
              <a href="/login" className="text-link">
                Back to Login
              </a>
            </div>
          </>
        )}
      </div>
    </div>
  );
};

export default VerifyOTP;
