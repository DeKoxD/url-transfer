function logger(elLog) {
	this.elLog = elLog;
}

logger.prototype.log = function() {
	if(!this.elLog) {
		return;
	}
	for(var i = 0; i < arguments.length; i++) {
		const p = document.createElement('pre');
		p.innerText = new Date().toISOString() + ": ";
		try {
			p.innerText += JSON.stringify(arguments[i], null, 2);
		} catch(error) {
			p.innerText = error + " -> " + arguments[i]; 
		}
		this.elLog.insertBefore(p, this.elLog.firstChild);
	}
}