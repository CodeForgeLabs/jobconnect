"use client"

const GetStarted = () => {
  return (
    <div className="hero bg-base-200 min-h-screen">
  <div className="hero-content flex-col lg:flex-row-reverse">
    <div className="text-center lg:text-left">
      <h1 className="text-5xl font-bold">Welcome to Job Connect!</h1>
      <p className="py-6">
        Find the right talent.
Or become the talent.
      Log in to connect with clients, showcase your skills, and build your freelance career.
Or hire talented professionals to bring your ideas to life.
      </p>
    </div>
    <div className="card bg-base-100 w-full max-w-sm shrink-0 shadow-2xl">
      <div className="card-body">
        <fieldset className="fieldset">
          <legend>Get Started</legend>
          <p>Join Job Connect today and unlock a world of opportunities. Whether you're a freelancer looking for exciting projects      
or a client seeking top talent, Job Connect is your gateway to success. Sign up now and start connecting with the right people to achieve your goals!</p>
          <button className="btn btn-neutral mt-4">Get Started</button>
        </fieldset>
      </div>
    </div>
  </div>
</div>
  )
}

export default GetStarted