import { createContext, useContext, useEffect, useRef } from 'react';
import WebSocketContext from '../WebSocketContext';
import { State } from '../enums.tsx'


class DrawCallback {
    constructor(
        public callback: (d: DrawData) => void = (_: DrawData) => {},
    ) {}
}
const DrawCallbackContext = createContext<DrawCallback>(new DrawCallback());

/*
// https://developer.mozilla.org/en-US/docs/Web/API/MouseEvent/buttons
enum MouseButtons {
    None = 0,
    Primary = 1,
    Secondary = 2,
    Auxiliary = 4,
    Back = 8,
    Forward = 16,
}
*/

// We could try to be clever and use window.getComputedStyle(document.body).getPropertyValue('--player-1')
// to ensure these don't go out of sync, but for now it's sufficient (and efficient) to just define twice:
const playerColors: string[] = [
    '#ff3232',
    '#ff9232',
    '#e7ff32',
    '#32ff87',
    '#32ffee',
    '#3295ff',
    '#c132ff',
    '#ff326f',
];

// https://developer.mozilla.org/en-US/docs/Web/API/MouseEvent/button
enum MouseButton {
    Primary = 0,
    Secondary = 1,
    Auxiliary = 2,
    Back = 3,
    Forward = 4,
}

class DrawData {
    constructor(
        public lastX: number,
        public lastY: number,
        public x: number,
        public y: number,
        public playerNumber: number,
    ) {}

    // Reset all coordinates to the same point.
    // Used to start a new stroke, allows a dot.
    reset(x: number, y: number) {
        this.lastX = x;
        this.lastY = y;
        this.x = x;
        this.y = y;
    }

    // Cycles the coordinates. Used to track an in-progress stroke.
    update(x: number, y: number) {
        this.lastX = this.x;
        this.lastY = this.y;
        this.x = x;
        this.y = y;
    }
}

interface CanvasProps {
    currentPlayer: number;
    gameState: State;
    playerNumber: number;
}

function Canvas(props: CanvasProps) {
    const canvasRef = useRef<HTMLCanvasElement>(null);
    const canvasWrapperRef = useRef<HTMLDivElement>(null);
    const ws = useContext(WebSocketContext);
    if (ws === null) {
        console.error("got a null websocket")
    }
    const drawCallback = useContext(DrawCallbackContext);

    const freeDraw = (props.gameState === State.Waiting);
    const playerTurn = (props.gameState === State.Drawing) && (props.playerNumber === props.currentPlayer);
    let canDraw = freeDraw || playerTurn;

    useEffect(() => {
        if (canvasRef.current === null) { return; }
        if (canvasWrapperRef.current === null) { return; }
        const canvas = canvasRef.current;
        const canvasWrapper = canvasWrapperRef.current;
        canvas.width = canvasWrapper.clientWidth;
        canvas.height = canvasWrapper.clientHeight;

        const context = canvas.getContext('2d');
        if (context === null) { return; }

        // Get a non-null context to avoid constantly checking later.
        const ctx: CanvasRenderingContext2D = context;

        canvas.style.background = 'white';
        ctx.fillStyle = 'rgba(0, 0, 0, 0.5)';

        /*
        function resize() {
            // Busted, clears image each time.
            // Using canvas.getImageData/putImageData doesn't re-scale.
            // Using an Image allows scaling, but it's lossy.
            canvas.width = window.innerWidth;
            canvas.height = window.innerHeight;
            ctx.clearRect(0,   0, canvas.width, canvas.height);
        }
        window.onresize = resize;
        resize()
        */
        //canvas.width = window.innerWidth;
        //canvas.height = window.innerHeight;

        // Clear canvas when room is created and when each game starts.
        if (props.gameState === State.Waiting || props.gameState === State.GettingPrompt) {
            ctx.clearRect(0, 0, canvas.width, canvas.height);
        }

        let drawing = false;
        let first = true;
        const drawData = new DrawData(0, 0, 0, 0, props.playerNumber);

        function draw(d: DrawData) {
            ctx.beginPath();
            ctx.moveTo(d.lastX, d.lastY);
            ctx.lineTo(d.x, d.y);

            ctx.strokeStyle = playerColors[d.playerNumber - 1]; // players 1-indexed, colors 0-indexed
            ctx.lineWidth = 5;
            ctx.lineCap = 'round';
            ctx.stroke();
            ctx.closePath();
        }
        drawCallback.callback = draw;

        function move(e: MouseEvent) {
            if (!canDraw) { return; }
            if (drawing) {
                if (first) {
                    drawData.reset(e.offsetX, e.offsetY);
                    first = false;
                }
                drawData.update(e.offsetX, e.offsetY);
                draw(drawData);
                if (ws !== null) {
                    ws.send(JSON.stringify({
                        type: 'draw',
                        data: drawData,
                    }));
                } else {
                    console.error("cannot send draw data: no WebSocket")
                }
            }
        }
        canvas.onmousemove = move;
        canvas.onmousedown = (e: MouseEvent) => {
            if (e.button != MouseButton.Primary) { return; }
            if (!canDraw) { return; }
            drawData.reset(e.offsetX, e.offsetY);
            first = true;
            drawing = true;
            draw(drawData);
        }
        canvas.onmouseup = (e: MouseEvent) => {
            if (e.button != MouseButton.Primary) { return; }
            if (!canDraw) { return; }
            // End turn if ending a stroke on current turn (NOT during Waiting)
            if (drawing && playerTurn) {
                // Note that this condition is important for protecting canDraw
                // in case drawing=false, canDraw=true, user clicks outside canvas,
                // then releases mouse inside canvas! This would waste the player's turn.

                // Prevent further drawing
                canDraw = false;
                // Notify server we're done.
                if (ws != null) {
                    ws.send(JSON.stringify({
                        type: 'done',
                        data: null,
                    }));
                } else {
                    console.error("cannot end turn: no WebSocket")
                }
            }
            first = true;
            drawing = false;
        }
        canvas.onmouseleave = (_: MouseEvent) => {
            // If we've left the canvas, stop drawing.
            first = true;
            drawing = false;
        }

        return () => { // cleanup
            window.onresize = null;
            drawCallback.callback = (_: DrawData) => { console.error("draw callback called after cleanup"); };
        };

    }, [props.playerNumber, props.gameState, props.currentPlayer, playerTurn, canDraw]);

    return (
        <div id="canvas-wrapper" ref={canvasWrapperRef}>
            <canvas ref={canvasRef}></canvas>
        </div>
    )
}

export { Canvas, DrawCallback, DrawCallbackContext };
