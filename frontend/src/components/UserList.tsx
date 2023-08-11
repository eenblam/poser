interface UserListProps {
  users: User[];
}

class User {
  constructor(
    public id: string,
    public name: string,
    public color: string,
    public owner: boolean,
  ) {}
}

function UserList(props: UserListProps) {
  const users = props.users;
  const listItems = users.map((user: User) =>
    <li key={user.id}>{user.id}</li>
  );
  return (
    <div>
      <h2>Users</h2>
      <ul>{listItems}</ul>
    </div>
  );
}

export { User, UserList }
