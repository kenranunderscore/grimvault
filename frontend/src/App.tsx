import { useState } from "react";
import "./App.css";

type Person = {
  id: number;
  name: string;
  active: boolean;
};

function debugPeople(people: Person[]) {
  people.forEach((p, i) => console.log(`${i}: ${p.name}, ${p.active}`));
}

function App() {
  const [people, setPeople] = useState<Person[]>([
    { id: 1, name: "Peter", active: false },
    { id: 2, name: "Paul", active: false },
    { id: 3, name: "Mary", active: false },
    { id: 4, name: "Marco", active: false },
  ]);

  const setPerson = (i: number, checked: boolean) => {
    setPeople((oldPeople) =>
      oldPeople.map((person, j) => (i === j ? { ...person, active: checked } : person)),
    );
  };

  return (
    <>
      <h1>GrimVault</h1>
      <ul>
        {people.map((p, i) => (
          <Item key={p.id} person={p} setChecked={(checked: boolean) => setPerson(i, checked)} />
        ))}
      </ul>
      <div className="card">
        <button onClick={() => debugPeople(people)}>click me you fool</button>
      </div>
    </>
  );
}

function Item({ person, setChecked }: { person: Person; setChecked: (checked: boolean) => void }) {
  return (
    <li>
      <input
        name="checker"
        type="checkbox"
        checked={person.active}
        onChange={() => setChecked(!person.active)}
      />
      <span>{person.name}</span>
    </li>
  );
}

export default App;
