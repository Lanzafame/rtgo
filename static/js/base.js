var Base = (function Base(global) {
    'use strict';

    var formContainer = clean('.form-container'),
        loginForm = clean('.form[name="login"]'),
        registerForm = clean('.form[name="register"]');

    rtgo.showRegister = function showRegister() {
        loginForm.addClass('hide');
        registerForm.remClass('hide');
        if (formContainer.hasClass('fade-down-paused')) {
            clean('.form-container').togClass('fade-down-paused', 'fade-down');
        } else if (formContainer.hasClass('fade-up')) {
            clean('.form-container').togClass('fade-up', 'fade-down');
        }
    };

    rtgo.showLogin = function showLogin() {
        registerForm.addClass('hide');
        loginForm.remClass('hide');
        if (formContainer.hasClass('fade-down-paused')) {
            clean('.form-container').togClass('fade-down-paused', 'fade-down');
        } else if (formContainer.hasClass('fade-up')) {
            clean('.form-container').togClass('fade-up', 'fade-down');
        }
    };

    rtgo.hideForms = function hideLogin() {
        clean('.form-container').togClass('fade-down', 'fade-up');
    };

}(window || this));
