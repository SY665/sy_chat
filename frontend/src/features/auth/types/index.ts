export interface AuthUser{
    id:string
    username:string
    email:string
    name:string
}

export interface AuthData{
    user:AuthUser
}

export interface RegisterInput{
    username:string
    email:string
    password:string
}

export interface LoginInput{
    identifier:string
    password:string
}

export interface LogoutData{
    message:string
}

