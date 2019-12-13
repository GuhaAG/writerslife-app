import React, { Component } from 'react';
import { Form, Row, Col } from 'react-bootstrap';
import Swal from 'sweetalert2';
import axios from 'axios';
import ErrorHandler from '../components/ErrorHandler';

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
            userId: null,
            username: "",
            firstname: "",
            lastname: "",
            email: "",
            nomdeplume: "",
            password: "",
            confirm_password: "",
            error: false,
            errorMessageProfile: "",
            errorMessageNomdeplume: "",
            errorMessagePassword: "",
        };

        this.handleChange = this.handleChange.bind(this);
    }

    componentDidMount() {
        axiosInstance.get('/users/')
            .then((response) => {
                this.setState({
                    userId: response.data.id,
                    username: response.data.username,
                    firstname: response.data.firstname,
                    lastname: response.data.lastname,
                    email: response.data.email,
                    nomdeplume: response.data.nomdeplume,
                    error: false
                });
            }, (error) => {
                ErrorHandler.sweetAlertApiErrorMessage(error);
            });
    }

    handleChange = event => {
        this.setState({
            [event.target.id]: event.target.value
        });
    }

    updateProfile = event => {
        event.preventDefault();

        var user = {
            id: this.state.userId,
            username: this.state.username,
            email: this.state.email,
            firstname: this.state.firstname,
            lastname: this.state.lastname
        }

        axiosInstance.put('/users/update/' + this.state.userId, user)
            .then(() => {
                this.setState({
                    error: false,
                    errorMessageProfile: ''
                });
                Swal.fire({
                    icon: 'success',
                    title: 'Success !',
                    html: 'Your profile was updated.'
                });
            }, (error) => {
                let errMessage = ErrorHandler.sweetAlertApiErrorMessage(error);
                this.setState({
                    error: true,
                    errorMessageProfile: errMessage
                });
            });
    }

    updateNomdeplume = event => {
        event.preventDefault();

        var user = {
            id: this.state.userId,
            nomdeplume: this.state.nomdeplume
        }

        axiosInstance.put('/users/update/' + this.state.userId, user)
            .then(() => {
                this.setState({
                    error: false,
                    errorMessageNomdeplume: ''
                });
                Swal.fire({
                    icon: 'success',
                    title: 'Your nomdeplume is saved in the annals of history',
                    html: 'You can use your nomdeplume to publish your writing.'
                });
            }, (error) => {
                let errMessage = ErrorHandler.sweetAlertApiErrorMessage(error);
                this.setState({
                    error: true,
                    errorMessageNomdeplume: errMessage
                });
            });
    }

    updatePassword = event => {
        event.preventDefault();

        if (this.state.password !== this.state.confirm_password) {
            this.setState({
                error: true,
                errorMessagePassword: "New passwords do not match"
            });
            Swal.fire({
                icon: 'error',
                title: '',
                text: "New passwords do not match",
            })

            return;
        }

        var user = {
            id: this.state.userId,
            password: this.state.password
        }

        axiosInstance.put('users/update/' + this.state.userId, user)
            .then(() => {
                this.setState({
                    error: false,
                    errorMessagePassword: ''
                });
                Swal.fire({
                    icon: 'success',
                    title: 'Success !',
                    html: 'Your password is updated'
                });
            }, (error) => {
                let errMessage = ErrorHandler.sweetAlertApiErrorMessage(error);

                this.setState({
                    error: true,
                    errorMessagePassword: errMessage
                });
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
                                <Form.Group as={Row} controlId="username">
                                    <Form.Label column sm="2">
                                        Username
                                    </Form.Label>
                                    <Col sm="10">
                                        <Form.Control type="text" value={this.state.username} onChange={this.handleChange} readOnly />
                                    </Col>
                                </Form.Group>
                                <Form.Group as={Row} controlId="email">
                                    <Form.Label column sm="2">
                                        Email
                                    </Form.Label>
                                    <Col sm="10">
                                        <Form.Control type="email" value={this.state.email} onChange={this.handleChange} />
                                    </Col>
                                </Form.Group>
                                <Form.Group as={Row} controlId="firstname">
                                    <Form.Label column sm="2">
                                        Firstname
                                    </Form.Label>
                                    <Col sm="10">
                                        <Form.Control type="text" value={this.state.firstname} onChange={this.handleChange} />
                                    </Col>
                                </Form.Group>
                                <Form.Group as={Row} controlId="lastname">
                                    <Form.Label column sm="2">
                                        Lastname
                                    </Form.Label>
                                    <Col sm="10">
                                        <Form.Control type="text" value={this.state.lastname} onChange={this.handleChange} />
                                    </Col>
                                </Form.Group>
                            </Form>
                            {this.state.error && <div>
                                <p className="text-red-500">{this.state.errorMessageProfile}</p>
                            </div>}

                            <button className="w-40 text-center py-3 rounded bg-blue-800 text-white focus:outline-none my-1" onClick={this.updateProfile}>
                                Update Profile
                            </button>
                        </div>
                        <div className="bg-white rounded shadow-md text-black w-full font-mono p-4">
                            <h1 className="mb-8 text-3xl text-center">Use a Nickname</h1>
                            <Form>
                                <Form.Group as={Row} controlId="nomdeplume">
                                    <Form.Label column sm="2">
                                        Nomdeplume
                                    </Form.Label>
                                    <Col sm="10">
                                        <Form.Control type="text" value={this.state.nomdeplume || ''} onChange={this.handleChange} />
                                    </Col>
                                </Form.Group>
                            </Form>
                            {this.state.error && <div>
                                <p className="text-red-500">{this.state.errorMessageNomdeplume}</p>
                            </div>}

                            <button className="w-40 text-center py-3 rounded bg-blue-800 text-white focus:outline-none my-1" onClick={this.updateNomdeplume}>
                                Save Nickname
                            </button>
                        </div>
                        <div className="bg-white rounded shadow-md text-black w-full font-mono p-4">
                            <h1 className="mb-8 text-3xl text-center">Change Password</h1>
                            <Form>
                                <Form.Group as={Row} controlId="password">
                                    <Form.Label column sm="2">
                                        New Password
                                    </Form.Label>
                                    <Col sm="10">
                                        <Form.Control type="password" onChange={this.handleChange} />
                                    </Col>
                                </Form.Group>
                                <Form.Group as={Row} controlId="confirm_password">
                                    <Form.Label column sm="2">
                                        Confirm New Password
                                    </Form.Label>
                                    <Col sm="10">
                                        <Form.Control type="password" onChange={this.handleChange} />
                                    </Col>
                                </Form.Group>
                            </Form>
                            {this.state.error && <div>
                                <p className="text-red-500">{this.state.errorMessagePassword}</p>
                            </div>}

                            <button className="w-40 text-center py-3 rounded bg-blue-800 text-white focus:outline-none my-1" onClick={this.updatePassword}>
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
