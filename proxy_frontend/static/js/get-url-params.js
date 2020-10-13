// Function from: https://gist.github.com/cvan/38fa77f1f28d3eb9d9c461e1d0d0d7d7
function getURLParams() {
	return window.location.search.slice(1).split('&').reduce(function (q, query) {
		var chunks = query.split('=');
		var key = chunks[0];
		var value = decodeURIComponent(chunks[1]);
		value = isNaN(Number(value))? value : Number(value);
		return (q[key] = value, q);
	}, {});
};