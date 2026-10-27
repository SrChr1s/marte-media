export namespace main {
	
	export class ResolvePrefs {
	    height: number;
	    videoCodecs: string[];
	    audioCodecs: string[];
	
	    static createFrom(source: any = {}) {
	        return new ResolvePrefs(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.height = source["height"];
	        this.videoCodecs = source["videoCodecs"];
	        this.audioCodecs = source["audioCodecs"];
	    }
	}

}

export namespace player {
	
	export class SourceInfo {
	    url: string;
	    mime: string;
	    height: number;
	    codec: string;
	
	    static createFrom(source: any = {}) {
	        return new SourceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.mime = source["mime"];
	        this.height = source["height"];
	        this.codec = source["codec"];
	    }
	}
	export class StreamInfo {
	    id: string;
	    title: string;
	    channel: string;
	    artist: string;
	    track: string;
	    thumbnail: string;
	    duration: number;
	    heights: number[];
	    video?: SourceInfo;
	    audio?: SourceInfo;
	    muxed?: SourceInfo;
	
	    static createFrom(source: any = {}) {
	        return new StreamInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.channel = source["channel"];
	        this.artist = source["artist"];
	        this.track = source["track"];
	        this.thumbnail = source["thumbnail"];
	        this.duration = source["duration"];
	        this.heights = source["heights"];
	        this.video = this.convertValues(source["video"], SourceInfo);
	        this.audio = this.convertValues(source["audio"], SourceInfo);
	        this.muxed = this.convertValues(source["muxed"], SourceInfo);
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

}

export namespace queue {
	
	export class Request {
	    url: string;
	    title: string;
	    channel: string;
	    thumbnail: string;
	    mode: string;
	    height: number;
	    videoFormat: string;
	    audioFormat: string;
	    audioBitrate: number;
	    dir: string;
	
	    static createFrom(source: any = {}) {
	        return new Request(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.title = source["title"];
	        this.channel = source["channel"];
	        this.thumbnail = source["thumbnail"];
	        this.mode = source["mode"];
	        this.height = source["height"];
	        this.videoFormat = source["videoFormat"];
	        this.audioFormat = source["audioFormat"];
	        this.audioBitrate = source["audioBitrate"];
	        this.dir = source["dir"];
	    }
	}
	export class Job {
	    id: string;
	    request: Request;
	    status: string;
	    percent: number;
	    downloaded: number;
	    total: number;
	    speed: number;
	    eta: number;
	    path: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new Job(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.request = this.convertValues(source["request"], Request);
	        this.status = source["status"];
	        this.percent = source["percent"];
	        this.downloaded = source["downloaded"];
	        this.total = source["total"];
	        this.speed = source["speed"];
	        this.eta = source["eta"];
	        this.path = source["path"];
	        this.error = source["error"];
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

}

export namespace store {
	
	export class HistoryEntry {
	    id: string;
	    url: string;
	    title: string;
	    channel: string;
	    thumbnail: string;
	    mode: string;
	    format: string;
	    path: string;
	    // Go type: time
	    finishedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new HistoryEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.url = source["url"];
	        this.title = source["title"];
	        this.channel = source["channel"];
	        this.thumbnail = source["thumbnail"];
	        this.mode = source["mode"];
	        this.format = source["format"];
	        this.path = source["path"];
	        this.finishedAt = this.convertValues(source["finishedAt"], null);
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
	export class Settings {
	    downloadDir: string;
	    maxConcurrent: number;
	    mode: string;
	    videoHeight: number;
	    videoFormat: string;
	    audioFormat: string;
	    audioBitrate: number;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.downloadDir = source["downloadDir"];
	        this.maxConcurrent = source["maxConcurrent"];
	        this.mode = source["mode"];
	        this.videoHeight = source["videoHeight"];
	        this.videoFormat = source["videoFormat"];
	        this.audioFormat = source["audioFormat"];
	        this.audioBitrate = source["audioBitrate"];
	    }
	}

}

export namespace ytdlp {
	
	export class Entry {
	    id: string;
	    url: string;
	    title: string;
	    channel: string;
	    duration: number;
	    thumbnail: string;
	    viewCount: number;
	
	    static createFrom(source: any = {}) {
	        return new Entry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.url = source["url"];
	        this.title = source["title"];
	        this.channel = source["channel"];
	        this.duration = source["duration"];
	        this.thumbnail = source["thumbnail"];
	        this.viewCount = source["viewCount"];
	    }
	}
	export class MediaInfo {
	    kind: string;
	    id: string;
	    url: string;
	    title: string;
	    channel: string;
	    duration: number;
	    thumbnail: string;
	    viewCount: number;
	    heights: number[];
	    entries: Entry[];
	    playlistCount: number;
	
	    static createFrom(source: any = {}) {
	        return new MediaInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.id = source["id"];
	        this.url = source["url"];
	        this.title = source["title"];
	        this.channel = source["channel"];
	        this.duration = source["duration"];
	        this.thumbnail = source["thumbnail"];
	        this.viewCount = source["viewCount"];
	        this.heights = source["heights"];
	        this.entries = this.convertValues(source["entries"], Entry);
	        this.playlistCount = source["playlistCount"];
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
	export class Status {
	    installed: boolean;
	    path: string;
	    version: string;
	    ffmpeg: string;
	    jsRuntime: string;
	    jsPath: string;
	
	    static createFrom(source: any = {}) {
	        return new Status(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.installed = source["installed"];
	        this.path = source["path"];
	        this.version = source["version"];
	        this.ffmpeg = source["ffmpeg"];
	        this.jsRuntime = source["jsRuntime"];
	        this.jsPath = source["jsPath"];
	    }
	}

}

