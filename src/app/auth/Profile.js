import React, { Component } from 'react';
import { Form, Row, Col } from 'react-bootstrap';
import Swal from 'sweetalert2';
import axios from 'axios';

const token = window.localStorage.getItem("loginJwt");
const axiosInstance = axios.create({
    baseURL: 'http://localhost:8081/',
    timeout: 1000,
    headers: {
        'Authorization': `Bearer ${token}`
    }
});

class Profile extends Component {

    constructor(props) {
        super(props);

        this.state = {
            username: "",
            firstname: "",
            lastname: "",
            email: "",
            nomdeplume: "Richard Bachman",
            password: "",
            confirm_password: "",
            error: false,
            errorMessage: ""
        };
    }

    componentDidMount() {
        axiosInstance.get('/users/')
            .then((response) => {
                this.setState({
                    username: response.data.username,
                    firstname: response.data.firstname,
                    lastname: response.data.lastname,
                    email: response.data.email,
                    nomdeplume: response.data.nomdeplume,
                    error: false,
                    errorMessage: ''
                });
            }, (error) => {
                let responseErrorMessage = "";
                if (error.response && error.response.data) {
                    responseErrorMessage = error.response.data.message;
                    if (error.response.data.errors && error.response.data.errors[0].defaultMessage) {
                        responseErrorMessage = error.response.data.errors[0].defaultMessage;
                    }
                }
                if (responseErrorMessage !== "") {
                    this.setState({
                        error: true,
                        errorMessage: responseErrorMessage
                    });
                    Swal.fire({
                        icon: 'error',
                        title: '',
                        text: responseErrorMessage,
                    })
                }
                else {
                    console.log(error);
                }
            });
    }

    render() {

        return (
            <div className="bg-grey-lighter flex flex-col">
                <form onSubmit={this.handleSubmit}>
                    <div className="container max-w-2xl mx-auto flex-1 flex flex-col items-center justify-center px-2">
                        <div className="bg-white rounded shadow-md text-black w-full font-mono p-4">
                            <h1 className="mb-8 text-3xl text-center">My Profile</h1>
                            <Form>
                                <Form.Group as={Row} controlId="formPlaintextEmail">
                                    <Form.Label column sm="2">
                                        Email
                                    </Form.Label>
                                    <Col sm="10">
                                        <Form.Control type="email" placeholder={this.state.email} />
                                    </Col>
                                </Form.Group>
                                <Form.Group as={Row} controlId="formPlaintextUsername">
                                    <Form.Label column sm="2">
                                        Username
                                    </Form.Label>
                                    <Col sm="10">
                                        <Form.Control type="text" placeholder={this.state.username} />
                                    </Col>
                                </Form.Group>
                                <Form.Group as={Row} controlId="formPlaintextFirstname">
                                    <Form.Label column sm="2">
                                        Firstname
                                    </Form.Label>
                                    <Col sm="10">
                                        <Form.Control type="text" placeholder={this.state.firstname} />
                                    </Col>
                                </Form.Group>
                                <Form.Group as={Row} controlId="formPlaintextLastname">
                                    <Form.Label column sm="2">
                                        Lastname
                                    </Form.Label>
                                    <Col sm="10">
                                        <Form.Control type="text" placeholder={this.state.lastname} />
                                    </Col>
                                </Form.Group>
                            </Form>
                            {this.state.error && <div>
                                <p className="text-red-500">{this.state.errorMessage}</p>
                            </div>}

                            <button className="w-40 text-center py-3 rounded bg-blue-800 text-white focus:outline-none my-1">
                                Update Profile
                            </button>
                        </div>
                        <div className="bg-white rounded shadow-md text-black w-full font-mono p-4">
                            <h1 className="mb-8 text-3xl text-center">Use a Nickname</h1>
                            <Form>
                                <Form.Group as={Row} controlId="formPlaintextNomdeplume">
                                    <Form.Label column sm="2">
                                        Nomdeplume
                                    </Form.Label>
                                    <Col sm="10">
                                        <Form.Control type="text" placeholder={this.state.nomdeplume} />
                                    </Col>
                                </Form.Group>
                            </Form>
                            {this.state.error && <div>
                                <p className="text-red-500">{this.state.errorMessage}</p>
                            </div>}

                            <button className="w-40 text-center py-3 rounded bg-blue-800 text-white focus:outline-none my-1">
                                Save Nickname
                            </button>
                        </div>
                        <div className="bg-white rounded shadow-md text-black w-full font-mono p-4">
                            <h1 className="mb-8 text-3xl text-center">Change Password</h1>
                            <Form>
                                <Form.Group as={Row} controlId="formPlaintextCurrentPassword">
                                    <Form.Label column sm="2">
                                        Current Password
                                    </Form.Label>
                                    <Col sm="10">
                                        <Form.Control type="password" />
                                    </Col>
                                </Form.Group>
                                <Form.Group as={Row} controlId="formPlaintextNewPassword">
                                    <Form.Label column sm="2">
                                        New Password
                                    </Form.Label>
                                    <Col sm="10">
                                        <Form.Control type="password" />
                                    </Col>
                                </Form.Group>
                                <Form.Group as={Row} controlId="formPlaintextConfirmNewPassword">
                                    <Form.Label column sm="2">
                                        Confirm New Password
                                    </Form.Label>
                                    <Col sm="10">
                                        <Form.Control type="password" />
                                    </Col>
                                </Form.Group>
                            </Form>
                            {this.state.error && <div>
                                <p className="text-red-500">{this.state.errorMessage}</p>
                            </div>}

                            <button className="w-40 text-center py-3 rounded bg-blue-800 text-white focus:outline-none my-1">
                                Save Password
                            </button>
                        </div>
                    </div>
                </form>
            </div>
        );
    }
}

export default Profile;
