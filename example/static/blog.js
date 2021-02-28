document.addEventListener("DOMContentLoaded", function() {
    if (document.getElementById("post-list")) {
        var blogList = new List('post-list', {
            valueNames: ['title'],
            page: 10,
            pagination: true
        });
    }
});
