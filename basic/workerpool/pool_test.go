package workerpool

import (
	"fmt"
	"gopkg.in/go-playground/pool.v3"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

type PoolPerson struct {
	Id    int
	Name  string
	Hobby []string

	workUnit pool.WorkUnit
	doneCh   chan interface{}
}

func (p PoolPerson) String() string {
	return fmt.Sprintf("Person{Id:%d, Name:%s}", p.Id, p.Name)
}

func TestWorkerPool(t *testing.T) {
	doneCh := make(chan interface{})
	var persons []*PoolPerson
	for i := 0; i < 10; i++ {
		persons = append(persons, &PoolPerson{Id: i, Name: "user" + strconv.Itoa(i), doneCh: doneCh})
	}
	p := pool.NewLimited(3)
	defer p.Close()

	for _, person := range persons {
		person.workUnit = p.Queue(getPersonHobbies(person))
	}

	ticker := time.NewTimer(5 * time.Second)
	remain := len(persons)
loop:
	for {
		select {
		case v := <-doneCh:
			if v == nil {
				remain--
			} else {
				p.Close()
				break loop
			}

			if remain == 0 {
				break loop
			}
		case <-ticker.C:
			fmt.Println("timeout!!")
			break loop
		}
	}
	fmt.Println("## Remain :", remain)
	//Start to get person Person{Id:1, Name:user1}'s hobby. slee: 3 secs
	//Start to get person Person{Id:5, Name:user5}'s hobby. slee: 1 secs
	//Start to get person Person{Id:2, Name:user2}'s hobby. slee: 3 secs
	//Start to get person Person{Id:3, Name:user3}'s hobby. slee: 3 secs
	//Start to get person Person{Id:4, Name:user4}'s hobby. slee: 2 secs
	//Start to get person Person{Id:7, Name:user7}'s hobby. slee: 1 secs
	//Start to get person Person{Id:6, Name:user6}'s hobby. slee: 2 secs
	//Start to get person Person{Id:9, Name:user9}'s hobby. slee: 3 secs
	//timeout!!
	//## Remain : 5
}

func getPersonHobbies(person *PoolPerson) pool.WorkFunc {
	return func(wu pool.WorkUnit) (interface{}, error) {
		sleepSec := rand.Intn(3) + 1
		fmt.Printf("Start to get person %v's hobby. slee: %d secs\n", person, sleepSec)

		time.Sleep(time.Duration(sleepSec) * time.Second)
		if wu.IsCancelled() {
			fmt.Printf("Persons[%s]'s work is cancelled\n", person)
			// return values not used
			return nil, nil
		}

		// ready for processing...
		switch person.Id {
		case 1:
			person.Hobby = []string{"hobby1"}
		case 2:
			person.Hobby = []string{"hobby2"}
		case 3:
			person.Hobby = []string{"hobby3"}
		case 4:
			person.Hobby = []string{"hobby4"}
		case 5:
			person.Hobby = []string{"hobby5"}
		}

		//if rand.Intn(10) == 1 {
		//	person.doneCh <- errors.New("force err")
		//} else {
		//	person.doneCh <- nil
		//}
		person.doneCh <- nil
		return nil, nil
	}
}
