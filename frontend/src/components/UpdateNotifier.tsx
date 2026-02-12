import { useEffect, useState } from 'react';
import { CheckYtDlpUpdate } from '../../wailsjs/go/main/App';
import { BrowserOpenURL } from '../../wailsjs/runtime';

interface UpdateInfo {
    local_version: string;
    latest_version: string;
    update_url: string;
    has_update: boolean;
}

export default function UpdateNotifier() {
    const [updateInfo, setUpdateInfo] = useState<UpdateInfo | null>(null);
    const [visible, setVisible] = useState(false);

    useEffect(() => {
        // Delay check slightly to not interfere with startup diagnostics
        const timer = setTimeout(() => {
            CheckYtDlpUpdate()
                .then((info: UpdateInfo) => {
                    if (info && info.has_update) {
                        setUpdateInfo(info);
                        setVisible(true);
                    }
                })
                .catch((err: Error) => console.error('Failed to check for yt-dlp updates:', err));
        }, 3000);

        return () => clearTimeout(timer);
    }, []);

    if (!visible || !updateInfo) return null;

    return (
        <div className="fixed bottom-6 right-6 z-50 animate-in fade-in slide-in-from-right-8 duration-500">
            <div className="bg-slate-800 border border-blue-500/30 rounded-xl p-4 shadow-2xl shadow-blue-900/20 max-w-sm flex flex-col gap-3">
                <div className="flex items-start justify-between gap-4">
                    <div className="flex-1">
                        <h4 className="text-blue-400 font-semibold text-sm mb-1">yt-dlp Update Available</h4>
                        <p className="text-slate-300 text-xs leading-relaxed">
                            New version <span className="text-white font-mono font-bold">{updateInfo.latest_version}</span> is available (Current: {updateInfo.local_version}).
                            To ensure download success, it is recommended to update from the official website.
                        </p>
                    </div>
                    <button 
                        onClick={() => setVisible(false)}
                        className="text-slate-500 hover:text-slate-300 transition-colors"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
                    </button>
                </div>
                <div className="flex justify-end gap-2">
                    <button
                        onClick={() => BrowserOpenURL(updateInfo.update_url)}
                        className="bg-blue-600 hover:bg-blue-500 text-white text-xs font-medium px-4 py-2 rounded-lg transition-colors"
                    >
                        Download Now
                    </button>
                </div>
            </div>
        </div>
    );
}
