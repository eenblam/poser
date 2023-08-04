import { useEffect, useRef } from 'react';

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

function Canvas() {
    const canvasRef = useRef<HTMLCanvasElement>(null);

    useEffect(() => {
        if (canvasRef.current === null) { return; }
        const canvas = canvasRef.current;

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
        ctx.clearRect(0,   0, canvas.width, canvas.height);

        let drawing = false;
        let first = true;
        const drawData = new DrawData(0, 0, 0, 0);

        function draw(d: DrawData) {
                ctx.beginPath();
                ctx.moveTo(d.lastX, d.lastY);
                ctx.lineTo(d.x, d.y);

                ctx.strokeStyle = 'pink'; //TODO player color
                ctx.lineWidth = 5;
                ctx.lineCap = 'round';
                ctx.stroke();
                ctx.closePath();
        }

        function move(e: MouseEvent) {
            if (drawing) {
                if (first) {
                    drawData.reset(e.offsetX, e.offsetY);
                    first = false;
                }
                drawData.update(e.offsetX, e.offsetY);
                draw(drawData);
                //TODO send(drawData);
            }
        }
        canvas.onmousemove = move;
        canvas.onmousedown = (e: MouseEvent) => {
            if (e.button != MouseButton.Primary) { return; }
            drawData.reset(e.offsetX, e.offsetY);
            first = true;
            drawing = true;
            draw(drawData);
        }
        canvas.onmouseup = (e: MouseEvent) => {
            if (e.button != MouseButton.Primary) { return; }
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
        };

    }, []);

    return (
        <canvas ref={canvasRef}></canvas>
    )
}

export default Canvas
