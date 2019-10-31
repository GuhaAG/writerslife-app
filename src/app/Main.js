import React from 'react';
import Signup from './auth/Signup';
import Login from './auth/Login';
import { Switch, BrowserRouter as Router, Route } from 'react-router-dom';

function Main() {
    return (
        <Router>
            <Switch>
                <Route path='/signup' component={Signup} />
                <Route path='/login' component={Login} />
            </Switch>
        </Router>
    );
}

export default Main;