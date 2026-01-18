export namespace config {
	
	export class Config {
	    vault_path: string;
	    model_path: string;
	    llm_model: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.vault_path = source["vault_path"];
	        this.model_path = source["model_path"];
	        this.llm_model = source["llm_model"];
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

