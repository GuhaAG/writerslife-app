import Swal from 'sweetalert2';

function sweetAlertApiErrorMessage(error) {
    let responseErrorMessage = "";
    if (error.response && error.response.data) {
        responseErrorMessage = error.response.data.message;
        if (error.response.data.errors && error.response.data.errors[0].defaultMessage) {
            responseErrorMessage = error.response.data.errors[0].defaultMessage;
        }
    }
    if (responseErrorMessage !== "") {
        Swal.fire({
            icon: 'error',
            title: '',
            text: responseErrorMessage,
        })
    }
    else {
        console.log(error);
    }
    return responseErrorMessage;
}

module.exports = {
    sweetAlertApiErrorMessage: sweetAlertApiErrorMessage
}