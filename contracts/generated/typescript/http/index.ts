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
    "/health/ready": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get: operations["getReadiness"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/rooms": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        post: operations["createRoom"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/rooms/{code}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get: operations["previewRoom"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/rooms/{code}/join": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        post: operations["joinRoom"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/rooms/{code}/resume": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        post: operations["resumeRoom"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/rooms/{code}/leave": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        post: operations["leaveRoom"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/replays/{matchId}/capabilities": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        post: operations["createReplayCapability"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/replays/{replayId}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get: operations["getReplay"];
        put?: never;
        post?: never;
        delete: operations["deleteReplay"];
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/solo/puzzles": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        post: operations["createSoloAssignment"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/solo/attempts/{attemptId}/hint": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        post: operations["requestSoloHint"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/solo/attempts/{attemptId}/complete": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        post: operations["validateSoloCompletion"];
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
        Readiness: {
            /** @enum {string} */
            status: "ready" | "not_ready";
            reason?: string;
            version?: string;
        };
        /** @enum {string} */
        SoloPlayStyle: "Guided" | "Classic";
        /** @enum {string} */
        SoloDifficulty: "Easy" | "Medium" | "Hard" | "Expert" | "Random";
        SoloAssignmentRequest: {
            difficulty: components["schemas"]["SoloDifficulty"];
            playStyle: components["schemas"]["SoloPlayStyle"];
            recentPuzzleIds: components["schemas"]["UUIDv7"][];
        };
        SoloAssignmentResponse: {
            attemptId: components["schemas"]["UUIDv7"];
            assignmentProof: string;
            clues: string;
            puzzleId: components["schemas"]["UUIDv7"];
            revision: number;
            /** @enum {string} */
            difficulty: "Easy" | "Medium" | "Hard" | "Expert";
            generatorVersion: string;
            solverVersion: string;
            transformationVersion: string;
            transformationSeed: components["schemas"]["SafeInteger"];
            issuedAtMs: components["schemas"]["SafeInteger"];
        };
        SoloBoardRequest: {
            assignmentProof: string;
            values: string;
        };
        SoloHintRequest: {
            assignmentProof: string;
            values: string;
            /** @enum {string} */
            level: "Nudge" | "Reveal";
        };
        SoloHintResponse: {
            /** @enum {string} */
            level: "Nudge" | "Reveal";
            /** @constant */
            penaltyMs: 20000;
            technique?: string;
            unitKind?: string;
            unitIndex?: number;
            affectedCells?: number[];
            cell?: number;
            value?: number;
        };
        SoloCompletionResponse: {
            complete: boolean;
            incorrectCells?: number[];
        };
        SuccessEnvelope: {
            requestId: components["schemas"]["UUIDv7"];
            version: components["schemas"]["SafeInteger"];
            data: {
                displayName: string;
            };
        };
        /** @enum {string} */
        RoomMode: "Coop";
        /** @enum {string} */
        RoomDifficulty: "Easy" | "Medium" | "Hard" | "Expert";
        /** @enum {string} */
        RoomErrorPreset: "Casual" | "Challenge" | "Blind" | "Clean";
        /** @enum {string} */
        RoomRole: "Player" | "Spectator";
        /** @enum {string} */
        RoomState: "Lobby" | "Countdown" | "InMatch" | "Results";
        RoomSettings: {
            mode: components["schemas"]["RoomMode"];
            difficulty: components["schemas"]["RoomDifficulty"];
            errorPreset: components["schemas"]["RoomErrorPreset"];
            hintsEnabled: boolean;
            sharedNotes: boolean;
            autoRemoveNotes: boolean;
            spectatorsAllowed: boolean;
        };
        RoomParticipant: {
            id: components["schemas"]["UUIDv7"];
            name: string;
            role: components["schemas"]["RoomRole"];
            isHost: boolean;
            isReady: boolean;
            joinedAt: components["schemas"]["SafeInteger"];
        };
        Room: {
            id: components["schemas"]["UUIDv7"];
            code: string;
            state: components["schemas"]["RoomState"];
            version: components["schemas"]["SafeInteger"];
            settings: components["schemas"]["RoomSettings"];
            hostId: components["schemas"]["UUIDv7"] | null;
            currentMatchId: components["schemas"]["UUIDv7"] | null;
            rematchNumber: components["schemas"]["SafeInteger"];
            participants: components["schemas"]["RoomParticipant"][];
        };
        Countdown: {
            matchId: components["schemas"]["UUIDv7"];
            generation: components["schemas"]["SafeInteger"];
            deadlineAt: components["schemas"]["SafeInteger"];
        };
        RoomResponse: {
            room: components["schemas"]["Room"];
            countdown?: components["schemas"]["Countdown"] | null;
            participants: components["schemas"]["RoomParticipant"][];
            self: {
                participantId: components["schemas"]["UUIDv7"];
            };
        };
        RoomPreviewResponse: {
            mode: components["schemas"]["RoomMode"];
            difficulty: components["schemas"]["RoomDifficulty"];
            state: components["schemas"]["RoomState"];
            locked: boolean;
            playerSeatsTotal: number;
            playerSeatsAvailable: number;
            spectatorSeatsTotal: number;
            spectatorSeatsAvailable: number;
        };
        CreateRoomRequest: {
            displayName: string;
            mode: components["schemas"]["RoomMode"];
            difficulty: components["schemas"]["RoomDifficulty"];
        };
        JoinRoomRequest: {
            displayName: string;
            role?: components["schemas"]["RoomRole"];
        };
        LeaveRoomRequest: {
            /** @enum {string} */
            intent?: "leave_lobby" | "become_spectator";
        };
        ReplayCapabilityResponse: {
            replayId: components["schemas"]["UUIDv7"];
            capability: string;
            shareUrl: string;
            expiresAt: components["schemas"]["SafeInteger"];
        };
        ErrorEnvelope: {
            error: {
                code: string;
                messageKey: string;
                requestId: string;
                retryable?: boolean;
                details: {
                    [key: string]: string | boolean | number | null;
                };
            };
        };
    };
    responses: {
        /** @description Invalid request shape */
        BadRequest: {
            headers: {
                [name: string]: unknown;
            };
            content: {
                "application/json": components["schemas"]["ErrorEnvelope"];
            };
        };
        /** @description Invalid or missing session */
        Unauthorized: {
            headers: {
                [name: string]: unknown;
            };
            content: {
                "application/json": components["schemas"]["ErrorEnvelope"];
            };
        };
        /** @description Room or resource not found */
        NotFound: {
            headers: {
                [name: string]: unknown;
            };
            content: {
                "application/json": components["schemas"]["ErrorEnvelope"];
            };
        };
        /** @description State conflict or active session exists */
        Conflict: {
            headers: {
                [name: string]: unknown;
            };
            content: {
                "application/json": components["schemas"]["ErrorEnvelope"];
            };
        };
        /** @description Domain validation failure */
        Unprocessable: {
            headers: {
                [name: string]: unknown;
            };
            content: {
                "application/json": components["schemas"]["ErrorEnvelope"];
            };
        };
    };
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
    getReadiness: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Database is migrated and ready */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Readiness"];
                };
            };
        };
    };
    createRoom: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["CreateRoomRequest"];
            };
        };
        responses: {
            /** @description Room created */
            201: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["RoomResponse"];
                };
            };
            400: components["responses"]["BadRequest"];
            409: components["responses"]["Conflict"];
            422: components["responses"]["Unprocessable"];
        };
    };
    previewRoom: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                code: string;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Room preview */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["RoomPreviewResponse"];
                };
            };
            404: components["responses"]["NotFound"];
        };
    };
    joinRoom: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                code: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["JoinRoomRequest"];
            };
        };
        responses: {
            /** @description Joined room */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["RoomResponse"];
                };
            };
            400: components["responses"]["BadRequest"];
            404: components["responses"]["NotFound"];
            409: components["responses"]["Conflict"];
            422: components["responses"]["Unprocessable"];
        };
    };
    resumeRoom: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                code: string;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Resumed session */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["RoomResponse"];
                };
            };
            401: components["responses"]["Unauthorized"];
            404: components["responses"]["NotFound"];
        };
    };
    leaveRoom: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                code: string;
            };
            cookie?: never;
        };
        requestBody?: {
            content: {
                "application/json": components["schemas"]["LeaveRoomRequest"];
            };
        };
        responses: {
            /** @description Left room */
            204: {
                headers: {
                    [name: string]: unknown;
                };
                content?: never;
            };
            400: components["responses"]["BadRequest"];
            401: components["responses"]["Unauthorized"];
            404: components["responses"]["NotFound"];
            422: components["responses"]["Unprocessable"];
        };
    };
    createReplayCapability: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                matchId: components["schemas"]["UUIDv7"];
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Replay read capability created */
            201: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ReplayCapabilityResponse"];
                };
            };
            404: components["responses"]["NotFound"];
        };
    };
    getReplay: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                replayId: components["schemas"]["UUIDv7"];
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Immutable public replay projection */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": Record<string, never>;
                };
            };
            404: components["responses"]["NotFound"];
        };
    };
    deleteReplay: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                replayId: components["schemas"]["UUIDv7"];
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Replay access deleted */
            204: {
                headers: {
                    [name: string]: unknown;
                };
                content?: never;
            };
            404: components["responses"]["NotFound"];
        };
    };
    createSoloAssignment: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["SoloAssignmentRequest"];
            };
        };
        responses: {
            /** @description Stateless signed Solo assignment */
            201: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SoloAssignmentResponse"];
                };
            };
            400: components["responses"]["BadRequest"];
        };
    };
    requestSoloHint: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                attemptId: components["schemas"]["UUIDv7"];
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["SoloHintRequest"];
            };
        };
        responses: {
            /** @description Deterministic Nudge or Reveal */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SoloHintResponse"];
                };
            };
            400: components["responses"]["BadRequest"];
            401: components["responses"]["Unauthorized"];
            422: components["responses"]["Unprocessable"];
        };
    };
    validateSoloCompletion: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                attemptId: components["schemas"]["UUIDv7"];
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["SoloBoardRequest"];
            };
        };
        responses: {
            /** @description Authoritative completion validation */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SoloCompletionResponse"];
                };
            };
            400: components["responses"]["BadRequest"];
            401: components["responses"]["Unauthorized"];
            422: components["responses"]["Unprocessable"];
        };
    };
}
