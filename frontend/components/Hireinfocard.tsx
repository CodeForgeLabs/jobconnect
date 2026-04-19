import Image, { type StaticImageData } from "next/image";
import type { ReactNode } from "react";

interface HireinfocardProps {
  icon: ReactNode | StaticImageData | string;
  title: string;
  description: string;
}

const Hireinfocard = ({ icon, title, description }: HireinfocardProps) => {
  const shouldRenderImage = typeof icon === "string" || (typeof icon === "object" && icon !== null && "src" in icon);

  return (
    <div className="card  bg-base-100 shadow-xl tablet:w-1/3">
        <div className="card-body">
            <div className="flex flex-col  gap-4">
              
                    {shouldRenderImage ? <Image src={icon as string | StaticImageData} alt={title} className="w-12 h-12" /> : icon}
              
             
                    <h2 className="card-title ">{title}</h2>
                    <p className="text-gray-500">{description}</p>
              
            </div>
        </div>

    </div>
  );
};

export default Hireinfocard;
