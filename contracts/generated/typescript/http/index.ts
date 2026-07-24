// Code generated from contracts. DO NOT EDIT.

export interface paths {
    "/health/live": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get: operations["getLiveness"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
}
export type webhooks = Record<string, never>;
export interface components {
    schemas: {
        UUIDv7: string;
        /** Format: int64 */
        SafeInteger: number;
        Liveness: {
            /** @constant */
            status: "live";
            version: string;
        };
        SuccessEnvelope: {
            requestId: components["schemas"]["UUIDv7"];
            version: components["schemas"]["SafeInteger"];
            data: {
                displayName: string;
            };
        };
        ErrorEnvelope: {
            error: {
                code: string;
                messageKey: string;
                requestId: components["schemas"]["UUIDv7"];
                retryable: boolean;
                details: {
                    [key: string]: string | boolean | number | null;
                };
            };
        };
    };
    responses: never;
    parameters: never;
    requestBodies: never;
    headers: never;
    pathItems: never;
}
export type $defs = Record<string, never>;
export interface operations {
    getLiveness: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Process is live */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Liveness"];
                };
            };
        };
    };
}
