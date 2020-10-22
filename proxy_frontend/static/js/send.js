var id;
var serverAddr = window.location.protocol + "//" + window.location.host;

var lggr = {};

var elH2ID = document.getElementById("id");
var elInputID = document.getElementById("id-input");


var elInputURL = document.getElementById("url-input");
var elButtonSendURL = document.getElementById("send-url-button");

var elPStatus = document.getElementById("status");

var elDivLog = document.getElementById('log');

const params = getURLParams();

if(params.logs && (params.logs !== "false" || params.logs !== "0")) {
	lggr = new logger(elDivLog);
	elDivLog.style.display = "block";
} else {
	lggr = { log: function() {} };
}

id = String(params.id);
if(!id || !id.length) {
	elInputID.style.display = 'block';
	elInputID.onkeyup = function(e) {
		id = e.target.value;
	}
} else {
	elH2ID.innerText = id;
	elH2ID.style.display = 'block';
}

var xhrURL = new XMLHttpRequest();
xhrURL.onload = function () {
	elPStatus.innerText = "Done!"
}
xhrURL.onerror = function() {
	elPStatus.innerText = "Error!\n" + xhrURL.response
}
xhrURL.onloadstart = function() {
	elPStatus.innerText = "Sending..."
}

function sendURL() {
	if(!id || !id.length) {
		return;
	}
	try {
		const payload = JSON.stringify({ url: elInputURL.value })
		xhrURL.open('PUT', serverAddr + '/url/' + id);
		xhrURL.setRequestHeader('Content-Type', 'application/json')
		xhrURL.send(payload);
		lggr.log("URL sent:", url);
	} catch (error) {
		lggr.log("Error sending URL", error)
	}
}

elButtonSendURL.onclick = sendURL;
elInputURL.onkeypress = function(e) {
	if(e.keyCode === 13) {
		sendURL()
	}
}