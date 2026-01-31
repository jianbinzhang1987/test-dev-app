export namespace internal {
	
	export class Node {
	    id: string;
	    name: string;
	    ip: string;
	    port: number;
	    protocol: string;
	    isMaster: boolean;
	    username?: string;
	    authMethod?: string;
	    keyPath?: string;
	
	    static createFrom(source: any = {}) {
	        return new Node(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.ip = source["ip"];
	        this.port = source["port"];
	        this.protocol = source["protocol"];
	        this.isMaster = source["isMaster"];
	        this.username = source["username"];
	        this.authMethod = source["authMethod"];
	        this.keyPath = source["keyPath"];
	    }
	}
	export class NodeStatus {
	    latency: number;
	    lastChecked: string;
	    status: string;
	    errorMsg: string;
	
	    static createFrom(source: any = {}) {
	        return new NodeStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.latency = source["latency"];
	        this.lastChecked = source["lastChecked"];
	        this.status = source["status"];
	        this.errorMsg = source["errorMsg"];
	    }
	}
	export class SVNResource {
	    id: string;
	    url: string;
	    name: string;
	    type: string;
	    revision: string;
	    status: string;
	    lastChecked: string;
	    size?: string;
	    username?: string;
	
	    static createFrom(source: any = {}) {
	        return new SVNResource(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.url = source["url"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.revision = source["revision"];
	        this.status = source["status"];
	        this.lastChecked = source["lastChecked"];
	        this.size = source["size"];
	        this.username = source["username"];
	    }
	}
	export class SVNTestResult {
	    ok: boolean;
	    revision?: string;
	    message?: string;
	    durationMs?: number;
	    checkedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new SVNTestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.revision = source["revision"];
	        this.message = source["message"];
	        this.durationMs = source["durationMs"];
	        this.checkedAt = source["checkedAt"];
	    }
	}
	export class TaskDefinition {
	    id: string;
	    name: string;
	    svnResourceId: string;
	    masterServerId: string;
	    slaveServerIds: string[];
	    remotePath: string;
	    slaveRemotePath?: string;
	    slaveRemotePaths?: Record<string, string>;
	    commands: string[];
	    status: string;
	    progress: number;
	    createdAt: string;
	    updatedAt: string;
	    lastRunAt?: string;
	    templateId?: string;
	
	    static createFrom(source: any = {}) {
	        return new TaskDefinition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.svnResourceId = source["svnResourceId"];
	        this.masterServerId = source["masterServerId"];
	        this.slaveServerIds = source["slaveServerIds"];
	        this.remotePath = source["remotePath"];
	        this.slaveRemotePath = source["slaveRemotePath"];
	        this.slaveRemotePaths = source["slaveRemotePaths"];
	        this.commands = source["commands"];
	        this.status = source["status"];
	        this.progress = source["progress"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.lastRunAt = source["lastRunAt"];
	        this.templateId = source["templateId"];
	    }
	}
	export class TaskRun {
	    id: string;
	    taskId: string;
	    taskName: string;
	    status: string;
	    progress: number;
	    startedAt: string;
	    finishedAt?: string;
	    logs: string[];
	
	    static createFrom(source: any = {}) {
	        return new TaskRun(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.taskId = source["taskId"];
	        this.taskName = source["taskName"];
	        this.status = source["status"];
	        this.progress = source["progress"];
	        this.startedAt = source["startedAt"];
	        this.finishedAt = source["finishedAt"];
	        this.logs = source["logs"];
	    }
	}
	export class TaskRunRequest {
	    taskId: string;
	    taskName?: string;
	    svnResourceId: string;
	    masterServerId: string;
	    slaveServerIds: string[];
	    remotePath: string;
	    slaveRemotePath?: string;
	    slaveRemotePaths?: Record<string, string>;
	    commands: string[];
	
	    static createFrom(source: any = {}) {
	        return new TaskRunRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.taskId = source["taskId"];
	        this.taskName = source["taskName"];
	        this.svnResourceId = source["svnResourceId"];
	        this.masterServerId = source["masterServerId"];
	        this.slaveServerIds = source["slaveServerIds"];
	        this.remotePath = source["remotePath"];
	        this.slaveRemotePath = source["slaveRemotePath"];
	        this.slaveRemotePaths = source["slaveRemotePaths"];
	        this.commands = source["commands"];
	    }
	}
	export class TaskTemplate {
	    id: string;
	    name: string;
	    svnResourceId: string;
	    masterServerId: string;
	    slaveServerIds: string[];
	    remotePath: string;
	    slaveRemotePath?: string;
	    slaveRemotePaths?: Record<string, string>;
	    commands: string[];
	    sourceTaskId?: string;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new TaskTemplate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.svnResourceId = source["svnResourceId"];
	        this.masterServerId = source["masterServerId"];
	        this.slaveServerIds = source["slaveServerIds"];
	        this.remotePath = source["remotePath"];
	        this.slaveRemotePath = source["slaveRemotePath"];
	        this.slaveRemotePaths = source["slaveRemotePaths"];
	        this.commands = source["commands"];
	        this.sourceTaskId = source["sourceTaskId"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class TopologyData {
	    master?: Node;
	    slaves: Node[];
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new TopologyData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.master = this.convertValues(source["master"], Node);
	        this.slaves = this.convertValues(source["slaves"], Node);
	        this.total = source["total"];
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

