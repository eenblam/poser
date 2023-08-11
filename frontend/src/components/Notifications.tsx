interface NotificationsProps {
  notifications: Notification[];
}

class Notification {
  constructor(
    public timestamp: number,
    public message: string,
    public isError: boolean,
  ) {}
}

function Notifications(props: NotificationsProps) {
    // Leave sorting to websocket receiver, since it can insert to maintain continuous sort.
    let listItems = props.notifications.map((n: Notification) =>
                                            <p key={n.timestamp} className={n.isError ? "notification-error" : "notification"}>
                                                {n.message}
                                            </p>
                                        );
    return (
        <div id="notifications-div">
            {listItems}
        </div>
    );

}

export { Notification, Notifications }
