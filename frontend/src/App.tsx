import {useState} from 'react';
import Dashboard from './Dashboard';
import Settings from './Settings';
import './App.css';

function App() {
    const [view, setView] = useState('dashboard');

    return (
        <div className="flex flex-col h-screen bg-slate-900 text-slate-100 font-sans">
            {/* Top Navigation Bar */}
            <div className="flex items-center justify-center gap-6 py-4 border-b border-slate-800 bg-slate-900/50 backdrop-blur-sm sticky top-0 z-10">
                <button 
                    className={`px-4 py-2 text-sm font-medium rounded-full transition-all duration-200 ${
                        view === 'dashboard' 
                        ? 'bg-blue-600 text-white shadow-lg shadow-blue-900/20' 
                        : 'text-slate-400 hover:text-white hover:bg-slate-800'
                    }`} 
                    onClick={() => setView('dashboard')}
                >
                    Task
                </button>
                <button 
                    className={`px-4 py-2 text-sm font-medium rounded-full transition-all duration-200 ${
                        view === 'settings' 
                        ? 'bg-blue-600 text-white shadow-lg shadow-blue-900/20' 
                        : 'text-slate-400 hover:text-white hover:bg-slate-800'
                    }`} 
                    onClick={() => setView('settings')}
                >
                    Settings
                </button>
            </div>

            {/* Main Content Area */}
            <main className="flex-1 overflow-hidden flex flex-col relative">
                <div className={`absolute inset-0 overflow-y-auto ${view === 'dashboard' ? 'block' : 'hidden'}`}>
                    <Dashboard />
                </div>
                <div className={`absolute inset-0 overflow-y-auto ${view === 'settings' ? 'block' : 'hidden'}`}>
                    <Settings />
                </div>
            </main>
        </div>
    )
}
export default App;
