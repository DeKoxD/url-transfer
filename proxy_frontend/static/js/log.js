function logger(elLog) {
	this.elLog = elLog;
}

logger.prototype.log = function() {
	if(!this.elLog) {
		return;
	}
	const p = document.createElement('pre');
	p.textContent = new Date().toISOString() + ":\n";
	for(var i = 0; i < arguments.length; i++) {
		try {
			p.textContent += JSON.stringify(arguments[i], null, 2) + "\n";
		} catch(error) {
			p.textContent += error + " -> " + arguments[i] + "\n"; 
		}
	}
	this.elLog.insertBefore(p, this.elLog.firstChild);
}