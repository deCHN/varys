import { main } from "../../wailsjs/go/models";
import { useEffect, useState } from "react";

interface StartupHealthWizardProps {
    diagnostics: main.StartupDiagnostics | null;
    open: boolean;
    onClose: () => void;
    onRecheck: () => Promise<void>;
    onOpenSettings: () => void;
    onStartOllama: () => Promise<void>;
    onStopOllama: () => Promise<void>;
    onOpenModelLibrary: () => Promise<void>;
    onBrowseVaultPath: () => Promise<void>;
    onBrowseModelPath: () => Promise<void>;
    onPasteOpenAIKey: (key?: string) => Promise<string>;
}

function statusClass(status: string): string {
    if (status === "ok") {
        return "text-emerald-400 bg-emerald-500/10 border-emerald-500/30";
    }
    if (status === "missing") {
        return "text-red-400 bg-red-500/10 border-red-500/30";
    }
    return "text-amber-400 bg-amber-500/10 border-amber-500/30";
}

async function copyToClipboard(text: string): Promise<void> {
    if (navigator.clipboard && navigator.clipboard.writeText) {
        await navigator.clipboard.writeText(text);
        return;
    }

    const textarea = document.createElement("textarea");
    textarea.value = text;
    document.body.appendChild(textarea);
    textarea.select();
    document.execCommand("copy");
    document.body.removeChild(textarea);
}

function isURL(text: string): boolean {
    return /^https?:\/\//.test(text);
}

export default function StartupHealthWizard(props: StartupHealthWizardProps) {
    const { diagnostics, open, onClose, onRecheck, onOpenSettings } = props;
    const [maskedOpenAIKey, setMaskedOpenAIKey] = useState("");

    useEffect(() => {
        if (!diagnostics) {
            return;
        }
        const openAIItem = diagnostics.items.find((item) => item.id === "openai_key");
        if (openAIItem?.detected_path) {
            setMaskedOpenAIKey(openAIItem.detected_path);
        }
    }, [diagnostics]);

    if (!open || !diagnostics) {
        return null;
    }

    const maskSecret = (value: string): string => {
        const text = (value || "").trim();
        if (!text) {
            return "";
        }
        if (text.length <= 8) {
            return text;
        }
        return `${text.slice(0, 4)}${"*".repeat(text.length - 8)}${text.slice(-4)}`;
    };

    const pasteOpenAIKey = async () => {
        try {
            const text = (await props.onPasteOpenAIKey()).trim();
            if (!text) {
                return;
            }
            setMaskedOpenAIKey(maskSecret(text));
        } catch (err) {
            console.error("Failed to paste OpenAI key", err);
        }
    };

    const handleOpenAIKeyPaste = async (event: React.ClipboardEvent<HTMLInputElement>) => {
        event.preventDefault();
        const pasted = event.clipboardData.getData("text").trim();
        if (!pasted) {
            return;
        }
        try {
            const saved = (await props.onPasteOpenAIKey(pasted)).trim();
            setMaskedOpenAIKey(maskSecret(saved || pasted));
        } catch (err) {
            console.error("Failed to paste OpenAI key from input", err);
        }
    };

    return (
        <div className="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm flex items-center justify-center p-4">
            <div className="w-full max-w-4xl max-h-[90vh] overflow-hidden bg-slate-900 border border-slate-700 rounded-xl shadow-2xl">
                <div className="px-6 py-4 border-b border-slate-800 flex items-center justify-between">
                    <div>
                        <h2 className="text-xl font-semibold text-slate-100">Startup Dependency Health</h2>
                        <p className="text-sm text-slate-400 mt-1">
                            Provider: <span className="font-mono">{diagnostics.provider}</span>
                            {diagnostics.ready ? " · Ready" : ` · Blocked (${diagnostics.blockers.length})`}
                        </p>
                    </div>
                    <button
                        onClick={onClose}
                        className="px-3 py-1.5 text-sm rounded-lg bg-slate-800 text-slate-300 hover:bg-slate-700"
                    >
                        Close
                    </button>
                </div>

                <div className="p-6 space-y-4 overflow-y-auto max-h-[68vh]">
                    {diagnostics.items.map((item) => (
                        <div key={item.id} className="border border-slate-800 rounded-lg p-4 bg-slate-900/60">
                            <div className="flex items-center justify-between gap-3">
                                <div>
                                    <div className="text-slate-100 font-medium">{item.name}</div>
                                    <div className="text-xs text-slate-500 mt-1">ID: {item.id}</div>
                                </div>
                                <div className={`text-xs px-2 py-1 border rounded-md ${statusClass(item.status)}`}>
                                    {item.status.toUpperCase()}
                                    {item.is_blocker ? " · BLOCKER" : ""}
                                </div>
                            </div>

                            {item.status === "ok" ? null : (
                                <div className="mt-3 text-sm text-slate-300">{item.fix_suggestion}</div>
                            )}

                            {item.id !== "openai_key" && item.detected_path ? (
                                <div className="mt-2 text-xs text-slate-400 font-mono break-all">Detected: {item.detected_path}</div>
                            ) : null}

                            {item.status !== "ok" && item.id !== "openai_key" && item.fix_commands && item.fix_commands.length > 0 ? (
                                <div className="mt-3 space-y-2">
                                    {item.fix_commands.map((cmd, idx) => (
                                        <div key={`${item.id}-${idx}`} className="flex items-center gap-2">
                                            <code className="flex-1 text-xs text-slate-200 bg-slate-800 border border-slate-700 rounded-md px-2 py-1 break-all">
                                                {cmd}
                                            </code>
                                            {isURL(cmd) ? (
                                                <button
                                                    onClick={() => window.open(cmd, "_blank", "noopener,noreferrer")}
                                                    className="px-2 py-1 text-xs rounded-md bg-slate-700 text-white hover:bg-slate-600"
                                                >
                                                    Open
                                                </button>
                                            ) : null}
                                            {item.id !== "vault_path" && item.id !== "model_path" ? (
                                                <button
                                                    onClick={() => copyToClipboard(cmd)}
                                                    className="px-2 py-1 text-xs rounded-md bg-blue-600 text-white hover:bg-blue-500"
                                                >
                                                    Copy
                                                </button>
                                            ) : null}
                                        </div>
                                    ))}
                                </div>
                            ) : null}

                            {item.status !== "ok" && item.id === "vault_path" ? (
                                <div className="mt-3">
                                    <button
                                        onClick={props.onBrowseVaultPath}
                                        className="px-2 py-1 text-xs rounded-md bg-blue-600 text-white hover:bg-blue-500"
                                    >
                                        Browse Vault Path
                                    </button>
                                </div>
                            ) : null}

                            {item.status !== "ok" && item.id === "model_path" ? (
                                <div className="mt-3">
                                    <button
                                        onClick={props.onBrowseModelPath}
                                        className="px-2 py-1 text-xs rounded-md bg-blue-600 text-white hover:bg-blue-500"
                                    >
                                        Browse Whisper Model
                                    </button>
                                </div>
                            ) : null}

                            {item.id === "openai_key" ? (
                                <div className="mt-3 flex items-center gap-2">
                                    <input
                                        className="flex-1 text-xs text-slate-200 bg-slate-800 border border-slate-700 rounded-md px-2 py-1"
                                        placeholder="Paste OpenAI API Key from clipboard"
                                        value={maskedOpenAIKey}
                                        onPaste={handleOpenAIKeyPaste}
                                        onChange={() => {}}
                                    />
                                    <button
                                        onClick={pasteOpenAIKey}
                                        className="px-2 py-1 text-xs rounded-md bg-blue-600 text-white hover:bg-blue-500"
                                    >
                                        Paste
                                    </button>
                                </div>
                            ) : null}

                            {item.id === "ollama" ? (
                                <div className="mt-3">
                                    <button
                                        onClick={item.status === "ok" ? props.onStopOllama : props.onStartOllama}
                                        className={`px-2 py-1 text-xs rounded-md text-white ${
                                            item.status === "ok"
                                                ? "bg-red-600 hover:bg-red-500"
                                                : "bg-emerald-600 hover:bg-emerald-500"
                                        }`}
                                    >
                                        {item.status === "ok" ? "Stop Ollama" : "Start Ollama"}
                                    </button>
                                </div>
                            ) : null}

                            {item.status !== "ok" && item.id === "ollama_models" ? (
                                <div className="mt-3">
                                    <button
                                        onClick={props.onOpenModelLibrary}
                                        className="px-2 py-1 text-xs rounded-md bg-indigo-600 text-white hover:bg-indigo-500"
                                    >
                                        Open Model Download Page
                                    </button>
                                </div>
                            ) : null}
                        </div>
                    ))}
                </div>

                <div className="px-6 py-4 border-t border-slate-800 flex items-center justify-end gap-2">
                    <button
                        onClick={onOpenSettings}
                        className="px-3 py-2 text-sm rounded-lg bg-slate-700 text-slate-100 hover:bg-slate-600"
                    >
                        Open Settings
                    </button>
                    <button
                        onClick={onRecheck}
                        className="px-3 py-2 text-sm rounded-lg bg-blue-600 text-white hover:bg-blue-500"
                    >
                        Re-check
                    </button>
                </div>
            </div>
        </div>
    );
}
