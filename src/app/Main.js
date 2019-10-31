import React from 'react';
import Signup from './auth/Signup';
import { BrowserRouter as Router, Route } from 'react-router-dom';

function Main() {
    return (
        <Router>
            <Route path='/signup' component={Signup} />
        </Router>
    );
}

export default Main;