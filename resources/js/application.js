$(function() {

	$("#button_preview").on("click", function(e) {
		e.preventDefault();

		md = $("textarea").val();

		x = $.post("/markdown", md);
		x.done(function(html) {
			el = $("#preview");
			el.html(html);
			el.find("pre code").each(function(i, block) {
				hljs.highlightBlock(block);
			});
		});
		x.fail(function(res) {
			alert(res.responseText);
		});
	});
});