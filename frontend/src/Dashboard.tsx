import {useState, useEffect, useRef} from 'react';
import {SubmitTask} from "../wailsjs/go/main/App";
import {EventsOn} from "../wailsjs/runtime";

export default function Dashboard() {
    const [resultText, setResultText] = useState("");
    const [url, setUrl] = useState('');
    const [logs, setLogs] = useState<string[]>([]);
    const [analysisStream, setAnalysisStream] = useState("");
    const logEndRef = useRef<HTMLDivElement>(null);
    const streamEndRef = useRef<HTMLDivElement>(null);

    const updateUrl = (e: any) => setUrl(e.target.value);

    function addLog(msg: string) {
        // Simple timestamp
        const time = new Date().toLocaleTimeString([], { hour12: false });
        setLogs(prev => [...prev, `[${time}] ${msg}`]);
    }

    useEffect(() => {
        const unsubLog = EventsOn("task:log", (msg: string) => addLog(msg));
        const unsubAnalysis = EventsOn("task:analysis", (chunk: string) => {
            setAnalysisStream(prev => prev + chunk);
        });
        return () => {
            unsubLog();
            unsubAnalysis();
        };
    }, []);

    useEffect(() => {
        logEndRef.current?.scrollIntoView({ behavior: "smooth" });
    }, [logs]);

    useEffect(() => {
        streamEndRef.current?.scrollIntoView({ behavior: "smooth" });
    }, [analysisStream]);

    function processUrl() {
        if (!url) return;
        setLogs([]);
        setAnalysisStream("");
        setResultText("Processing...");
        addLog(`Processing URL: ${url}`);
        
        SubmitTask(url).then((response: string) => {
             addLog(`Backend Response: ${response}`);
             setResultText("Task completed");
        }).catch((err: any) => {
             addLog(`Error: ${err}`);
             setResultText("Task failed");
        });
    }

    return (
        <div style={{ display: 'flex', flexDirection: 'column', flex: 1, padding: '24px', maxWidth: '900px', margin: '0 auto', height: '100%' }}>
            <div style={{ display: 'flex', gap: 12, marginBottom: '24px' }}>
                <input 
                    className="input" 
                    value={url} 
                    onChange={updateUrl} 
                    placeholder="Enter YouTube/Bilibili URL"
                    style={{ flex: 1 }}
                />
                <button className="btn" onClick={processUrl}>Process</button>
            </div>

            <div style={{ display: 'flex', gap: '20px', flex: 1, overflow: 'hidden' }}>
                <div className="console-log" style={{ flex: 1 }}>
                    {logs.length === 0 && <div style={{ color: '#9aa0a6', fontStyle: 'italic' }}>Logs will appear here...</div>}
                    {logs.map((log, index) => (
                        <div key={index} style={{ marginBottom: '4px' }}>{log}</div>
                    ))}
                    <div ref={logEndRef} />
                </div>

                {analysisStream && (
                    <div className="console-log" style={{ flex: 1, borderLeft: '1px solid #ddd', paddingLeft: '12px' }}>
                        <div style={{ fontWeight: 600, marginBottom: '8px', color: '#1a73e8' }}>Live Analysis</div>
                        <div style={{ whiteSpace: 'pre-wrap', fontSize: '13px', lineHeight: '1.5' }}>
                            {analysisStream}
                        </div>
                        <div ref={streamEndRef} />
                    </div>
                )}
            </div>
            
            {resultText && (
                <div style={{ marginTop: '12px', fontSize: '13px', color: resultText.includes("failed") ? '#d93025' : '#188038', fontWeight: 500 }}>
                    {resultText}
                </div>
            )}
        </div>
    )
}