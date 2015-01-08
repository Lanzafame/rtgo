//    Title: base.js
//    Author: JD
//
//    This program is free software: you can redistribute it and/or modify
//    it under the terms of the GNU General Public License as published by
//    the Free Software Foundation, either version 3 of the License, or
//    (at your option) any later version.
//
//    This program is distributed in the hope that it will be useful,
//    but WITHOUT ANY WARRANTY; without even the implied warranty of
//    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//    GNU General Public License for more details.
//
//    You should have received a copy of the GNU General Public License
//    along with this program.  If not, see <http://www.gnu.org/licenses/>.

var Base = (function Base(global) {
    'use strict';

    var formContainer = clean('.form-container'),
        loginForm = clean('.form[name="login"]'),
        registerForm = clean('.form[name="register"]');

/**
 * rtgo.showRegister
 * Show the registration form.
 */
    rtgo.showRegister = function showRegister() {
        loginForm.addClass('hide');
        registerForm.remClass('hide');
        if (formContainer.hasClass('fade-down-paused')) {
            clean('.form-container').togClass('fade-down-paused', 'fade-down');
        } else if (formContainer.hasClass('fade-up')) {
            clean('.form-container').togClass('fade-up', 'fade-down');
        }
    };

/**
 * rtgo.showLogin
 * Show the login form.
 */
    rtgo.showLogin = function showLogin() {
        registerForm.addClass('hide');
        loginForm.remClass('hide');
        if (formContainer.hasClass('fade-down-paused')) {
            clean('.form-container').togClass('fade-down-paused', 'fade-down');
        } else if (formContainer.hasClass('fade-up')) {
            clean('.form-container').togClass('fade-up', 'fade-down');
        }
    };

/**
 * rtgo.hideForms
 * Hide all forms.
 */
    rtgo.hideForms = function hideLogin() {
        clean('.form-container').togClass('fade-down', 'fade-up');
    };

}(window || this));
