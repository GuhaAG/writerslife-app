import React from 'react';
import Home from './Home';
import Signup from './auth/Signup';
import Login from './auth/Login';
import Profile from './auth/Profile';
import Settings from './auth/Settings';
import PasswordReset from './auth/PasswordReset';
import PrivateRoute from './PrivateRoute';
import { Switch, BrowserRouter as Router, Route } from 'react-router-dom';

function Main() {
    return (
        <Router>
            <Switch>
                <PrivateRoute exact path='/' component={Home} />
                <Route path='/Signup' component={Signup} />
                <Route path='/login' component={Login} />
                <Route path='/PasswordReset' component={PasswordReset} />
                <PrivateRoute exact path='/Profile' component={Profile} />
                <PrivateRoute exact path='/Settings' component={Settings} />
            </Switch>
        </Router>
    );
}

export default Main;