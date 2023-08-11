interface UserListProps {
  users: User[];
}

class User {
  constructor(
    public id: string,
    public playerNumber: number,
    public name: string,
    public owner: boolean,
  ) {}
}

function UserList(props: UserListProps) {
  const users = props.users;
  const listItems = users.map((user: User) => {
    const className = `player-${user.playerNumber}`
    const userName = `Player #${user.playerNumber}`;
    return (<li key={user.playerNumber} className={className}>{userName}</li>);
  });
  return (
    <div>
      <h2>Users</h2>
      <ul>{listItems}</ul>
    </div>
  );
}

export { User, UserList }
