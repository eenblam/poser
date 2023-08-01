interface UserListProps {
  users: string[];
}

function UserList(props: UserListProps) {
  const users = props.users;
  const listItems = users.map((user: string) =>
    <li key={user}>{user}</li>
  );
  return (
    <div>
      <h2>Users</h2>
      <ul>{listItems}</ul>
    </div>
  );
}

export default UserList
