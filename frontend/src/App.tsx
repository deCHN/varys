import {useState} from 'react';
import Dashboard from './Dashboard';
import Settings from './Settings';
import './App.css';

function App() {
    const [view, setView] = useState('dashboard');

    return (
        <div id="App" style={{ display: 'flex', flexDirection: 'column', height: '100vh' }}>
            <div className="tab-nav">
                <button 
                    className={`tab-btn ${view === 'dashboard' ? 'active' : ''}`} 
                    onClick={() => setView('dashboard')}
                >
                    Task
                </button>
                <button 
                    className={`tab-btn ${view === 'settings' ? 'active' : ''}`} 
                    onClick={() => setView('settings')}
                >
                    Settings
                </button>
            </div>
            <main style={{ flex: 1, overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
                {view === 'dashboard' ? <Dashboard /> : <Settings />}
            </main>
        </div>
    )
}
export default App;
