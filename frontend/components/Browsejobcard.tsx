interface BrowsejobcardProps {
  image: string | { src: string };
  title: string;
  jobs: string;
}

const Browsejobcard = ({ image, title, jobs }: BrowsejobcardProps) => {
  const imageSrc = typeof image === "string" ? image : image.src;

  return (
    <div className="card h-72 shadow-xl tablet:w-1/4 relative overflow-hidden rounded-2xl border border-white/25 isolate">
      <div
        className=" absolute inset-0  bg-no-repeat bg-cover bg-local "
        style={{ backgroundImage: `url(${imageSrc})` }}
      />
      <div className="absolute inset-0 bg-black/35" />

      <div className="card-body pt-28 absolute left-0 bottom-0 z-10">
        <div className="flex flex-col gap-1 text-white">
          <h2 className="card-title">{title}</h2>
          <p className="text-gray-200">{jobs}</p>
        </div>
      </div>
    </div>
  );
};

export default Browsejobcard;
