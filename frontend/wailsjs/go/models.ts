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

}

