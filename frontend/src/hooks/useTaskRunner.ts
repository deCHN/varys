import { useState, useEffect, useCallback } from 'react';
import { SubmitTask, CancelTask } from "../../wailsjs/go/main/App";
import { EventsOn } from "../../wailsjs/runtime";

export function useTaskRunner() {
    const [isProcessing, setIsProcessing] = useState(false);
    const [logs, setLogs] = useState<string[]>([]);
    const [analysisStream, setAnalysisStream] = useState("");
    const [progress, setProgress] = useState(0);
    const [resultText, setResultText] = useState("");

    const addLog = useCallback((msg: string) => {
        const time = new Date().toLocaleTimeString([], { hour12: false });
        setLogs(prev => [...prev, `[${time}] ${msg}`]);
    }, []);

    useEffect(() => {
        const unsubLog = EventsOn("task:log", (msg: string) => addLog(msg));
        const unsubAnalysis = EventsOn("task:analysis", (chunk: string) => {
            setAnalysisStream(prev => prev + chunk);
        });
        const unsubProgress = EventsOn("task:progress", (p: number) => {
            setProgress(p);
        });

        return () => {
            unsubLog();
            unsubAnalysis();
            unsubProgress();
        };
    }, [addLog]);

    const runTask = async (url: string, downloadVideo: boolean) => {
        if (!url) return;
        
        setLogs([]);
        setAnalysisStream("");
        setProgress(0);
        setIsProcessing(true);
        setResultText("Processing...");
        
        const audioOnly = !downloadVideo;
        addLog(`Processing URL: ${url} (AudioOnly: ${audioOnly})`);

        try {
            const response = await SubmitTask(url, audioOnly);
            addLog(`Backend Response: ${response}`);
            setResultText("Task completed");
        } catch (err: any) {
            addLog(`Error: ${err}`);
            setResultText("Task failed");
        } finally {
            setProgress(0);
            setIsProcessing(false);
        }
    };

    const cancel = async () => {
        if (!isProcessing) return;
        try {
            await CancelTask();
            addLog("User requested cancellation...");
            setIsProcessing(false);
            setResultText("Cancelled");
            setProgress(0);
        } catch (err: any) {
            addLog(`Failed to cancel: ${err}`);
        }
    };

    return {
        isProcessing,
        logs,
        analysisStream,
        progress,
        resultText,
        runTask,
        cancel
    };
}
