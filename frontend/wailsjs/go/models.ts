export namespace config {
	
	export class Config {
	    vault_path: string;
	    model_path: string;
	    llm_model: string;
	    translation_model: string;
	    target_language: string;
	    context_size: number;
	    custom_prompt: string;
	    ai_provider: string;
	    openai_model: string;
	    openai_key: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.vault_path = source["vault_path"];
	        this.model_path = source["model_path"];
	        this.llm_model = source["llm_model"];
	        this.translation_model = source["translation_model"];
	        this.target_language = source["target_language"];
	        this.context_size = source["context_size"];
	        this.custom_prompt = source["custom_prompt"];
	        this.ai_provider = source["ai_provider"];
	        this.openai_model = source["openai_model"];
	        this.openai_key = source["openai_key"];
	    }
	}

}

export namespace main {
	
	export class DependencyStatus {
	    yt_dlp: boolean;
	    ffmpeg: boolean;
	    whisper: boolean;
	    ollama: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DependencyStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.yt_dlp = source["yt_dlp"];
	        this.ffmpeg = source["ffmpeg"];
	        this.whisper = source["whisper"];
	        this.ollama = source["ollama"];
	    }
	}
	export class DiagnosticItem {
	    id: string;
	    name: string;
	    status: string;
	    required_for: string[];
	    detected_path: string;
	    fix_suggestion: string;
	    fix_commands: string[];
	    can_auto_fix: boolean;
	    is_blocker: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DiagnosticItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.status = source["status"];
	        this.required_for = source["required_for"];
	        this.detected_path = source["detected_path"];
	        this.fix_suggestion = source["fix_suggestion"];
	        this.fix_commands = source["fix_commands"];
	        this.can_auto_fix = source["can_auto_fix"];
	        this.is_blocker = source["is_blocker"];
	    }
	}
	export class StartupDiagnostics {
	    generated_at: string;
	    provider: string;
	    blockers: string[];
	    ready: boolean;
	    items: DiagnosticItem[];
	
	    static createFrom(source: any = {}) {
	        return new StartupDiagnostics(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.generated_at = source["generated_at"];
	        this.provider = source["provider"];
	        this.blockers = source["blockers"];
	        this.ready = source["ready"];
	        this.items = this.convertValues(source["items"], DiagnosticItem);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class YtDlpUpdateInfo {
	    local_version: string;
	    latest_version: string;
	    update_url: string;
	    has_update: boolean;
	
	    static createFrom(source: any = {}) {
	        return new YtDlpUpdateInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.local_version = source["local_version"];
	        this.latest_version = source["latest_version"];
	        this.update_url = source["update_url"];
	        this.has_update = source["has_update"];
	    }
	}

}

