import React from 'react';

const TestPage: React.FC = () => {
  return (
    <div className="auth-container">
      <div className="auth-card">
        <h2>Test Page - CSS Check</h2>

        <div className="alert alert-success" style={{ marginBottom: '16px' }}>
          This is a success alert - CSS is working!
        </div>

        <div className="alert alert-error" style={{ marginBottom: '16px' }}>
          This is an error alert
        </div>

        <div className="alert alert-info" style={{ marginBottom: '16px' }}>
          This is an info alert
        </div>

        <form>
          <div className="form-group">
            <label htmlFor="test">Test Input</label>
            <input
              type="text"
              id="test"
              placeholder="Type something..."
            />
          </div>

          <button type="submit" className="btn btn-primary btn-block">
            Primary Button
          </button>

          <div style={{ marginTop: '16px' }}>
            <button type="button" className="btn btn-outline btn-block">
              Outline Button
            </button>
          </div>
        </form>

        <div className="auth-divider">DIVIDER</div>

        <div className="manual-entry">
          <p>Manual entry section:</p>
          <code className="secret-code">TESTCODE123456</code>
        </div>

        <div className="loading-container">
          <div className="spinner"></div>
          <p>Loading spinner test</p>
        </div>
      </div>
    </div>
  );
};

export default TestPage;
