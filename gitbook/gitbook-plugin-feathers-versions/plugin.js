require(['gitbook', 'jQuery'], function (gitbook, $) {
    var versions = [];

    // Update the select with a list of versions
    function updateVersions(url) {
        // Cleanup existing selector
        $('.versions-select').remove();
    
        if (versions.length == 0) return;

        var $li = $('<li>', {
            'class': 'versions-select',
            'html': '<div><select></select></div>'
        });
        var $select = $li.find('select');

        $.each(versions, function(i, version) {
            var $option = $('<option>', {
                'selected': version.value === window.location.origin,
                'value': version.value,
                'text': version.text
            });

            $option.appendTo($select);
        });

        $select.change(function() {
            var filtered = $.grep(versions, function(v) {
                return v.value === $select.val();
            });
            // Get actual version Object from array
            var version = filtered[0];
            window.location.href = version.value;
        });

        $li.prependTo('.book-summary ul.summary'); 
    }

    gitbook.events.bind('start', function (e, config) {
        var pluginConfig = config.versions || {};

        if (pluginConfig.gitbookConfigURL) {
            $.getJSON(pluginConfig.gitbookConfigURL).then(function(bookJSON) {
                if(bookJSON.pluginsConfig && bookJSON.pluginsConfig.versions &&
                    bookJSON.pluginsConfig.versions.options) {
                    versions = bookJSON.pluginsConfig.versions.options;

                    updateVersions();
                }
            });
        }
    });

    gitbook.events.bind('page.change', function () {
        updateVersions();
    });
});
