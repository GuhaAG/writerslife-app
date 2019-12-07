import React from 'react';
import Signup from './auth/Signup';
import Login from './auth/Login';
import Profile from './auth/Profile';
import Home from './Home';
import PrivateRoute from './PrivateRoute';
import { Switch, BrowserRouter as Router, Route } from 'react-router-dom';

function Main() {
    return (
        <Router>
            <Switch>
                <PrivateRoute exact path='/' component={Home} />
                <Route path='/signup' component={Signup} />
                <Route path='/login' component={Login} />
                <PrivateRoute exact path='/Profile' component={Profile} />
            </Switch>
        </Router>
    );
}

export default Main;