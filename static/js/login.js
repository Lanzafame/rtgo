var Login = (function Login(global) {
    "use strict";

    var submit = clean('.form-button');

/**
 * sendForm
 * Send the form data to the server via an XHR.
 * @param {Object} values
 */
    function sendForm(values) {
        var fd = new FormData();

        if (!values.username || !values.password || (values.type === 'register' && !values.email)) {
            return;
        }
        fd.append('username', values.username);
        fd.append('password', sjcl.codec.hex.fromBits(sjcl.hash.sha256.hash(values.password)));
        if (values.type === 'register') {
            fd.append('email', values.email);
        }
        clean.xhrReq({
            url: global.location.protocol + '//' + global.location.hostname + ':' + global.location.port + '/' + values.type,
            method: 'post',
            data: fd,
            success: function (e, xhr, response) {
                console.log('Login success: ' + response);
            },
            failure: function (e, xhr) {
                console.log('Login failed: ' + e);
            }
        });
    }

/**
 * checkFields
 * Check the form fields upon submitting the form.
 * If any of the validations fail, add a pulsing red
 * border animation to the input field.
 * @param {Event Object} e
 */
    function checkFields(e) {
        var wordregx = /^\w+$/,
            emailregx = /^[\w.%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,4}$/,
            form = e.target.getAttribute('data-form'),
            inputs = clean('form[name="' + form + '"] .form-input'),
            values = {
                type: form
            };
            
        inputs.eachnode(function (node) {
            var name = node.name,
                val = node.value;

            if (((name === 'username' || name === 'password') && !wordregx.test(val)) || (name === 'email' && !emailregx.test(val))) {
                val = '';
            }
            if (!val) {
                node.value = '';
                node.parentNode.classList.remove('fail');
                setTimeout(function () {
                    node.parentNode.classList.add('fail');
                }, 200);
            } else {
                values[name] = val;
                node.value = '';
            }
        });
        sendForm(values);
    }
    submit.on('click', checkFields, false);

/**
 * addGlow
 * Add a light-blue glow to the input field icons when in focus.
 * @param {Event Object} e
 */
    function addGlow(e) {
        var icon = e.target.parentNode.querySelector('.form-input-icon');

        icon.classList.add('glow');
    }

/**
 * removeGlow
 * Remove the light-blue glow from the input field icons when out of focus.
 * @param {Event Object} e
 */
    function removeGlow(e) {
        var icon = e.target.parentNode.querySelector('.form-input-icon');

        icon.classList.remove('glow');
    }
    clean('.form-input').on('focus', addGlow, false);
    clean('.form-input').on('blur', removeGlow, false);

}(this));
