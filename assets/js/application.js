import xhook from 'xhook';
import $ from 'jquery';
window.jQuery = $;
window.$ = $;
import 'popper.js';
import 'parsleyjs';
import 'bootstrap';

import '../scss/application.scss'

$(document).ready(function () {

    //make dropdown link navigatable
    $('.navbar .dropdown-toggle').click(function () {
        if (!isMobileDevice())
            window.location = $(this).attr('href');
    });

    if (document.querySelector('#ck-content')) {
        //add csrf protection to ckeditor uploads
        xhook.before(function (request) {
            if (!/^(GET|HEAD|OPTIONS|TRACE)$/i.test(request.method)) {
                request.xhr.setRequestHeader("X-CSRF-TOKEN", window.csrf_token);
            }
        });

        ClassicEditor
            .create(document.querySelector('#ck-content'), {
                language: 'ru', //to set different lang include <script src="/public/js/ckeditor/build/translations/{lang}.js"></script> along with core ckeditor script
                ckfinder: {
                    uploadUrl: '/admin/upload'
                },
            })
            .catch(error => {
                console.error(error);
            });
    }

    //initialize ckeditor on settings form page
    toggleCkEditor();
    window.toggleCkEditor = toggleCkEditor;
});

// Write your Javascript code.
function isMobileDevice() {
    return (typeof window.orientation !== "undefined") || (navigator.userAgent.indexOf('IEMobile') !== -1);
};

function toggleCkEditor() {
    var ctype = $('select[name="content_type"]');
    var ckConfig = {
        language: 'ru',
        allowedContent: true,
        filebrowserUploadUrl: '/admin/uploader/upload',
        extraPlugins: 'uploadimage,justify'
    };
    if (ctype && ctype.length > 0) {
        if (ctype.val() == 'html') {
            ClassicEditor
                .create(document.querySelector('#content'), {
                    language: 'ru',
                    ckfinder: {
                        uploadUrl: '/admin/upload'
                    },
                })
                .then(editor => {
                    window.ckeditor = editor;
                })
                .catch(error => {
                    console.error(error);
                });
        } else {
            //destroy ckeditor
            if (window.ckeditor) {
                console.log(2222, window.ckeditor);
                window.ckeditor.destroy();
            }
        }
    }
}