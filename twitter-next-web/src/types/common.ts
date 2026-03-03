export interface ValidationError {
    field: string;
    message: string;
}

/** Backend error payload. Use fieldErrors for form binding. */
export interface ErrorResponse {
    timestamp: string;
    status: number;
    error: string;
    message: string;
    path: string;
    errors?: ValidationError[];
}

/** Map of field name -> message for easy form input binding. */
export type FieldErrors = Record<string, string>;

export interface PageResponse<T> {
    content: T[];
    page: number;
    size: number;
    totalElements: number;
    totalPages: number;
    last: boolean;
}

export interface ApiResponse {
    success: boolean;
    message: string;
}
