document.addEventListener("DOMContentLoaded", function() {
    if (document.getElementById("post-list")) {
        const blogList = new List('post-list', {
            valueNames: [
                'title',
                'tag',
                {data: ['id'] }
            ],
            page: 10,
            pagination: true
        });

        // Add an onclick handler to filter by tag
        const tagSearches = document.querySelectorAll('.tag-search');
        tagSearches.forEach(function (tagList) {
            const tagElements = tagList.getElementsByTagName('li');
            for (let tagElement of tagElements) {
                tagElement.addEventListener('click', function () {
                    filterToTag(tagElement.innerHTML, blogList);
                }, false);
            }
        });
    }

});

function filterToTag(tagName, blogList) {
    // Remove the filter so all posts are shown
    blogList.filter()

    // Filter the list based on the selected tag
    if (tagName != "") {
        blogList.filter(function (item) {
            const itemNode = document.querySelector('[data-id="' + item.values().id + '"]');
            if (!itemNode) {
                console.log("Couldn't find list item " + item.values().id);
                return false;
            }

            const tagList = itemNode.querySelector('.tag-list');
            if (!tagList) {
                console.log("Couldn't find tag list for item " + item.values().id);
                return false;
            }

            matches = false;
            const tagElements = tagList.getElementsByTagName("li");
            for (let tagElement of tagElements) {
                if (tagElement.innerHTML == tagName) {
                    matches = true
                }
            }

            return matches
        });
    }
}
