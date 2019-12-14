import React, { Component } from 'react';
import axios from 'axios';
import Swal from 'sweetalert2';
import ErrorHandler from '../components/ErrorHandler';
import BlockUi from 'react-block-ui';
import 'react-block-ui/style.css';

const axiosInstance = axios.create({
    baseURL: 'http://localhost:8081/'
});

class PasswordRest extends Component {
    constructor(props) {
        super(props);

        this.state = {
            email: "",
            loading: false
        };
    }

    handleChange = event => {
        this.setState({
            [event.target.id]: event.target.value
        });
    }

    handleSubmit = event => {
        event.preventDefault();

        var emailUser = {
            email: this.state.email
        }

        this.setState({
            loading: true
        });

        axiosInstance.post('users/PasswordReset', emailUser)
            .then(() => {
                this.setState({
                    error: false,
                    errorMessage: '',
                    loading: false
                });
                let timerInterval

                Swal.fire({
                    icon: 'success',
                    title: 'Please check your registered email for our recovery email',
                    html: 'Taking you to Login page..',
                    timer: 5000,
                    timerProgressBar: true,
                    onBeforeOpen: () => {
                        Swal.showLoading();
                    },
                    onClose: () => {
                        clearInterval(timerInterval);
                        window.location.replace("/Login");
                    }
                }).then((result) => {
                    if (result.dismiss === Swal.DismissReason.timer) {
                        console.log('I was closed by the timer'); // eslint-disable-line
                    }
                })
            }, (error) => {
                let errMessage = ErrorHandler.sweetAlertApiErrorMessage(error);

                this.setState({
                    error: true,
                    errorMessage: errMessage,
                    loading: false
                });
            });
    }

    render() {

        return (

            <div className="bg-grey-lighter flex flex-col">
                <form onSubmit={this.handleSubmit}>
                    <BlockUi tag="div" blocking={this.state.loading}>
                        <div className="container max-w-sm mx-auto flex-1 flex flex-col items-center justify-center px-2">
                            <div className="bg-white px-6 py-8 rounded shadow-md text-black w-full font-mono">
                                <h1 className="mb-8 text-3xl text-center">Please enter your email address</h1>
                                <input
                                    required
                                    type="text"
                                    className="block border border-grey-light w-full p-3 rounded mb-4"
                                    id="email"
                                    onChange={this.handleChange}
                                    placeholder="Email" />

                                {this.state.error && <div>
                                    <p className="text-red-500">{this.state.errorMessage}</p>
                                </div>}

                                <button
                                    type="submit"
                                    className="w-full text-center py-3 rounded bg-blue-800 text-white focus:outline-none my-1"
                                >Password Reset</button>
                            </div>
                        </div>
                    </BlockUi>
                </form>
            </div>
        );
    }
}

export default PasswordRest;
