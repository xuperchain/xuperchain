#pragma once

namespace xchain {
enum class ErrorType {
    kSuccess = 0,
    kErrIteratorLoad,
    kErrTableIndexInvalid,
};

static const std::string errno_to_message(ErrorType no) {
    switch (no) {
        case ErrorType::kSuccess:
            return "success";
        case ErrorType::kErrIteratorLoad:
            return "iterator loading error";
        case ErrorType::kErrTableIndexInvalid:
            return "invalid index";
        default:
            return "unknown error";
    }
}

class Error {
public:
    Error(): _errno(ErrorType::kSuccess) {
        _message = errno_to_message(ErrorType::kSuccess);
    }

    Error(ErrorType no):  _errno(no) {
        _message = errno_to_message(no);
    }
    Error(ErrorType no, const std::string msg): _errno(no), _message(msg) {}

    bool operator()() {
        return ErrorType::kSuccess != _errno;
    }

    const std::string& operator()(bool) {
        return _message;
    }

private:
    ErrorType _errno;
    std::string _message;
};
} //end of xchain
