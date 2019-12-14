import React, { Component } from 'react';
import { Form } from 'react-bootstrap';

class Settings extends Component {

    constructor(props) {
        super(props);
        this.state = {
            userId: null,
            errMessage: ""
        };
    }

    render() {

        return (
            <div className="bg-grey-lighter flex flex-col">
                <form onSubmit={this.handleSubmit}>
                    <div className="container max-w-2xl mx-auto flex-1 flex flex-col items-center justify-center px-2">
                        <div className="bg-white rounded shadow-md text-black w-full font-mono p-4">
                            <h1 className="mb-8 text-3xl text-center">Settings</h1>
                            <Form>
                            </Form>
                        </div>
                    </div>
                </form>
            </div>
        );
    }
}

export default Settings;
